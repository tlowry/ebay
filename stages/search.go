package stages

import (
	"appengine"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"github.com/tlowry/grawl/browser"
	"github.com/tlowry/grawl/element"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type SearchStage struct {
	*pipeline.Stage
	term     string
	wg       *sync.WaitGroup
	Tier     string
	httpPool *util.Pool
}

func NewSearchStage(term string, wg *sync.WaitGroup, pool *util.Pool, ctx appengine.Context) *SearchStage {
	search := SearchStage{}
	search.Stage = pipeline.NewStage(ctx)
	search.term = term
	search.wg = wg
	search.Tier = "0"
	search.httpPool = pool

	return &search
}

func (search *SearchStage) Init() {
	search.SetName("SearchStage")
	search.Stage.Init()
}

func (search *SearchStage) MakeRequest(client *http.Client) {
	if client != nil {
		conn := browser.NewBrowserWithClient(client)
		conn.SetUserAgent("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.137 Safari/537.36")

		search.GetContext().Infof("Loading page")
		page := conn.Load("http://www.ebay.ie/sch/ebayadvsearch/")
		search.GetContext().Infof("Page Loaded")

		// Save the start time on page load to minimise inaccuracy
		startTime := time.Now()

		form := page.ById("adv_search_from").(*element.Form)
		if form == nil {
			search.GetContext().Errorf("searchStage::%s Couldn't find ebay search form", search.term)
		} else {
			form.ClearFields()

			form.SetField("_nkw", search.term)
			form.SetField("_in_kw", "1")
			form.SetField("_ex_kw", "")
			form.SetField("_sacat", "0")
			form.SetField("_udlo", "")
			form.SetField("_udhi", "")
			form.SetField("_ftrt", "901")
			form.SetField("_ftrv", "1")
			form.SetField("_sabdlo", "")
			form.SetField("_sabdhi", "")
			form.SetField("_samilow", "")
			form.SetField("_samihi", "0")
			form.SetField("_sadis", "10")
			form.SetField("_fpos", "")
			form.SetField("LH_SALE_CURRENCY", "0")
			form.SetField("_sop", "12")
			form.SetField("_dmd", "1")
			form.SetField("_ipg", "200")

			search.GetContext().Infof("Submitting form")
			page = conn.SubmitForm(form)
			search.GetContext().Infof("Form Submitted")
			//page.SaveToFile("./output/auctions-" + search.term + ".html")

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
				currency := prc.ByTag("b").GetContent()

				prcStr := util.SanitizeNum(prc.GetContent())
				price, err := strconv.ParseFloat(prcStr, 64)

				if err != nil {
					search.GetContext().Errorf("Error getting item price %s ", err.Error())
				}

				e := pipeline.EbayItem{}

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
				e.Currency = currency
				e.Url = link
				e.Price = price
				e.Tier = search.Tier

				search.GetContext().Infof("Search found ", e)
				search.Out <- e
				search.GetContext().Infof("Search passed on ", e)
				count++
			}
			search.GetContext().Infof("searchStage::", search.term, " found ", count, " items")
		}

	} else {
		search.GetContext().Infof("Failed to get http client")
	}
}

func (search *SearchStage) HandleIn() {
	search.GetContext().Infof("About to borrow a client")

	// Wait up to 10 seconds for a client
	cl := search.httpPool.BorrowWait()
	defer func() {
		e := recover()
		if e != nil {
			search.GetContext().Errorf("Hit an error in search stage %s", e)

		}
		if cl != nil {
			search.GetContext().Infof("About to return client")
			search.httpPool.Return(cl)
			search.GetContext().Infof("Client returned")
		} else {
			search.GetContext().Infof("Client Borrowed was nil")
		}
		search.wg.Done()

	}()

	if cl != nil {
		client := cl.(*http.Client)

		search.GetContext().Infof("Client Borrowed")
		search.MakeRequest(client)
		search.GetContext().Infof("Request completed normally")
	} else {
		search.GetContext().Infof("Failed to borrow a http client, received nil instead")
	}

}

func (search *SearchStage) Run() {
	search.HandleIn()
}
