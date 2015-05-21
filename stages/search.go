package stages

import (
	"appengine"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"github.com/tlowry/grawl/browser"
	"github.com/tlowry/grawl/element"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SearchStage struct {
	*pipeline.Stage
	term     string
	wg       *sync.WaitGroup
	Tier     string
	httpPool *util.Pool
	form     element.Form
	attempts int
}

func NewSearchStage(term string, wg *sync.WaitGroup, pool *util.Pool, form element.Form, ctx appengine.Context) *SearchStage {
	search := SearchStage{}
	search.Stage = pipeline.NewStage(ctx)
	search.term = term
	search.wg = wg
	search.Tier = "0"
	search.httpPool = pool
	search.form = form
	search.attempts = 0

	return &search
}

func (search *SearchStage) Init() {
	search.SetName("SearchStage")
	search.Stage.Init()
}
func (search *SearchStage) MakeRequest() *element.Page {

	search.GetContext().Infof("About to borrow a client")
	cl := search.httpPool.BorrowWait()
	search.GetContext().Infof("Client Borrowed")

	defer func() {
		if cl != nil {
			search.GetContext().Infof("About to return client")
			search.httpPool.Return(cl)
			search.GetContext().Infof("Client returned")
		} else {
			search.GetContext().Infof("Client Borrowed was nil")
		}
	}()

	var page *element.Page

	if cl != nil {
		client := cl.(*http.Client)

		if client != nil {
			conn := browser.NewBrowserWithClient(client)
			conn.SetUserAgent("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.137 Safari/537.36")

			form := search.form
			form.ClearFields()

			form.SetField("_nkw", search.term)
			form.SetField("_in_kw", "2")
			form.SetField("_ex_kw", "faulty")
			form.SetField("_sacat", "58058")
			form.SetField("_udlo", "")
			form.SetField("_udhi", "")
			form.SetField("_ftrt", "901")
			form.SetField("_ftrv", "2")
			form.SetField("_sabdlo", "")
			form.SetField("_sabdhi", "")
			form.SetField("_samilow", "")
			form.SetField("_samihi", "")
			form.SetField("_sadis", "10")
			form.SetField("_fpos", "")
			form.SetField("LH_Time", "1")
			form.SetField("LH_SALE_CURRENCY", "0")
			form.SetField("LH_TitleDesc", "1")
			form.SetField("_sop", "12")
			form.SetField("_dmd", "1")
			form.SetField("_ipg", "200")

			search.GetContext().Infof("Submitting form")
			page = conn.SubmitForm(&form)

			search.GetContext().Infof("Form Submitted")
			//page.SaveToFile("./output/auctions-" + search.term + ".html")

		} else {
			search.GetContext().Infof("Borrowed client has invalid type")
		}
	} else {
		search.GetContext().Infof("Client is nil")
	}

	return page
}

func (search *SearchStage) parsePage(page *element.Page) {
	// Save the start time on page load to minimise inaccuracy
	startTime := time.Now()
	// Only look in search results not related items
	results := page.ById("ResultSetItems")
	auctions := results.AllByClass("sresult")

	count := 0

	search.GetContext().Infof("Looking at auctions")
	for _, result := range auctions {

		listingId := result.GetAttribute("listingid")

		lnk := result.ByClass("img imgWr2")
		link := lnk.GetAttribute("href")

		img := lnk.ByTag("img")
		imgLink := img.GetAttribute("src")

		desc := img.GetAttribute("alt")

		prc := result.ByClass("lvprice prc").ByClass("bold")
		search.GetContext().Infof("%p ", prc)

		e := pipeline.EbayItem{}

		// Strike through/marked down prices
		if prc == nil {
			strikePrice := result.ByClass("stk-thr")
			if strikePrice != nil {
				currencyAndPrice := strings.Split(strikePrice.GetContent(), " ")
				if len(currencyAndPrice) > 1 {
					e.Currency = currencyAndPrice[0]
					prcStr := util.SanitizeNum(prc.GetContent())
					price, err := strconv.ParseFloat(prcStr, 64)
					if err == nil {
						e.Price = price

					} else {
						search.GetContext().Errorf("Error getting strikethrough item price %s ", err.Error())
					}
				}

			}

		}

		if prc != nil {
			cElem := prc.ByTag("b")
			currency := cElem.GetContent()
			e.Currency = currency

			prcStr := util.SanitizeNum(prc.GetContent())
			price, err := strconv.ParseFloat(prcStr, 64)
			if err == nil {
				e.Price = price

			} else {
				search.GetContext().Errorf("Error getting item price %s ", err.Error())
			}

		} else {
			search.GetContext().Infof("Failed to find price element ", prc)
		}

		// EndingTime
		endingTime := result.ByClass("timeMs")

		if endingTime != nil {
			search.GetContext().Infof("endingTime ok")
			timems := endingTime.GetAttribute("timems")
			timeStr := util.SanitizeNum(timems)

			timeMillis, timeErr := strconv.ParseInt(timeStr, 10, 64)

			if timeErr != nil {
				search.GetContext().Errorf("Error converting ending time", timeErr.Error())
			} else {
				millis := time.Duration(timeMillis)
				expiry := startTime.Add(time.Millisecond * millis)
				e.ExpiryDate = expiry
			}
		}

		e.Description = desc
		e.ListingId = listingId
		e.ImageUrl = imgLink
		e.Url = link

		e.Tier = search.Tier

		search.GetContext().Infof("Search found ", e)
		search.Out <- e
		search.GetContext().Infof("Search passed on ", e)
		count++
	}

	search.GetContext().Infof("searchStage::", search.term, " found ", count, " items")
}

func (search *SearchStage) MakeAttempts() {

	ctx := search.GetContext()

	defer func() {
		e := recover()

		if e != nil {
			err := e.(error)
			ctx.Infof("Error in search attempt %s", err.Error())
		}

		if search.attempts < 3 {
			search.attempts++
		}
	}()

	success := false

	for success == false && search.attempts < 3 {
		page := search.MakeRequest()
		search.GetContext().Infof("Request attempt %d ok, parsing response", search.attempts)
		search.parsePage(page)
		search.GetContext().Infof("Request attempt %d completed normally", search.attempts)
		success = true
	}

}
func (search *SearchStage) HandleIn() {

	defer func() {
		search.GetContext().Infof("Search complete for %s ", search.term)
		search.wg.Done()
	}()

	search.MakeAttempts()

}

func (search *SearchStage) Run() {
	search.HandleIn()
}
