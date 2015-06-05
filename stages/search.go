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
			search.GetContext().Errorf("Client Borrowed was nil")
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
			form.SetField("_in_kw", "1")
			form.SetField("_ex_kw", "faulty")
			form.SetField("_sacat", "58058")
			form.SetField("_udlo", "")
			form.SetField("_udhi", "")
			form.SetField("_ftrt", "901")
			form.SetField("_ftrv", "4")
			form.SetField("_sabdlo", "")
			form.SetField("_sabdhi", "")
			form.SetField("_samilow", "")
			form.SetField("_samihi", "")
			form.SetField("_sadis", "2000")
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

		} else {
			search.GetContext().Infof("Borrowed client has invalid type")
		}
	} else {
		search.GetContext().Infof("Client is nil")
	}

	return page
}

func (search *SearchStage) parsePrice(page element.Element) (price float64, currency string, err error) {
	price = 0
	currency = "€"

	prc := page.ByClass("lvprice prc").ByClass("bold")
	priceStr := ""

	if prc == nil {
		// Assume strike through/marked down prices
		strikePrice := page.ByClass("stk-thr")
		if strikePrice != nil {
			currencyAndPrice := strings.Split(strikePrice.GetContent(), " ")

			if len(currencyAndPrice) > 1 {
				currency = currencyAndPrice[0]
				priceStr = util.SanitizeNum(prc.GetContent())
			}
		}

	} else {
		cElem := prc.ByTag("b")
		currency = cElem.GetContent()
		priceStr = util.SanitizeNum(prc.GetContent())
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		search.GetContext().Errorf("Error getting item price %s ", err.Error())
	}

	return price, currency, err
}

func (search *SearchStage) parsePage(page *element.Page) {

	// Save the start time on page load to minimise inaccuracy
	startTime := time.Now()
	// Only look in search results not related items

	results := page.ById("ResultSetItems")
	count := 0
	if results != nil {
		auctions := results.AllByClass("sresult")

		for _, result := range auctions {

			listingId := result.GetAttribute("listingid")

			lnk := result.ByClass("img imgWr2")
			link := lnk.GetAttribute("href")

			img := lnk.ByTag("img")
			imgLink := img.GetAttribute("src")

			desc := img.GetAttribute("alt")

			e := pipeline.EbayItem{}

			var err error

			e.Price, e.Currency, err = search.parsePrice(result)

			if err != nil {
				search.GetContext().Errorf("Failed to parse price ", err)
			}

			// EndingTime
			endingTime := result.ByClass("timeMs")

			if endingTime != nil {
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

			search.Out <- e
			count++
		}
	} else {
		search.GetContext().Infof("Found no results for %s", search.term)
	}

	search.GetContext().Infof("searchStage::%s found %d items", search.term, count)
}

func (search *SearchStage) MakeAttempts() {

	ctx := search.GetContext()

	defer func() {
		e := recover()

		if e != nil {
			err := e.(error)
			ctx.Errorf("Error in search attempt %s", err)
		}

		if search.attempts < 2 {
			search.attempts++
		}
	}()

	success := false

	for success == false && search.attempts < 2 {
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
