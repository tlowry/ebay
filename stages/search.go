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

func (this *SearchStage) Init() {
	this.SetName("SearchStage")
	this.Stage.Init()
}

func (this *SearchStage) MakeRequest(client *http.Client) {
	if client != nil {
		conn := browser.NewBrowserWithClient(client)
		conn.SetUserAgent("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.137 Safari/537.36")

		/*
			page := conn.Load("http://www.ebay.ie/")
			form := page.ById("gh-f").(*element.Form)
			util.FailOnNil("searchStage::"+this.term+" Couldn't find ebay search form", form)

			form.SetField("_from", "R40")
			form.SetField("_trksid", "p2050601.m570.l1313")
			form.SetField("_nkw", this.term)

			// 200 results per page
			form.SetField("_ipg", "200")

			// category all
			form.SetField("_sacat", "0")

			// page number _pgn=2
		*/

		this.GetContext().Infof("Loading page")
		page := conn.Load("http://www.ebay.ie/sch/ebayadvsearch/")
		this.GetContext().Infof("Page Loaded")
		// Save the start time on page load to minimise inaccuracy
		startTime := time.Now()

		form := page.ById("adv_search_from").(*element.Form)
		util.FailOnNil("searchStage::"+this.term+" Couldn't find ebay search form", form)

		form.ClearFields()

		form.SetField("_nkw", this.term)
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
		form.SetField("_samihi", "")
		form.SetField("_sadis", "10")
		form.SetField("_fpos", "")
		form.SetField("LH_SALE_CURRENCY", "0")
		form.SetField("_sop", "12")
		form.SetField("_dmd", "1")
		form.SetField("_ipg", "200")

		this.GetContext().Infof("Submitting form")
		page = conn.SubmitForm(form)
		this.GetContext().Infof("Form Submitted")
		//page.SaveToFile("./output/auctions-" + this.term + ".html")

		// Only look in search results not related items
		results := page.ById("ResultSetItems")

		auctions := results.AllByClass("sresult")

		count := 0

		this.GetContext().Infof("Looking at auctions")
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
				this.GetContext().Errorf("Error getting item price %s ", err.Error())
			}

			e := pipeline.EbayItem{}

			// EndingTime
			endingTime := result.ByClass("timeMs")

			if endingTime != nil {
				this.GetContext().Infof("endingTime ok")
				timems := endingTime.GetAttribute("timems")
				timeStr := util.SanitizeNum(timems)

				timeMillis, timeErr := strconv.ParseInt(timeStr, 10, 64)

				if timeErr != nil {
					this.GetContext().Errorf("Error converting ending time", timeErr.Error())
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
			e.Tier = this.Tier

			this.GetContext().Infof("Search found ", e)
			this.Out <- e

			count++
		}
		this.GetContext().Infof("searchStage::", this.term, " found ", count, " items")
	} else {
		this.GetContext().Infof("Failed to get http client")
	}
}

func (this *SearchStage) HandleIn() {
	this.GetContext().Infof("About to borrow a client")

	cl := this.httpPool.Borrow(time.Second * 5)
	defer func() {
		e := recover()
		if e != nil {
			this.GetContext().Errorf("Hit an error in search stage %s", e)

		}
		if cl != nil {
			this.GetContext().Infof("About to return client")
			this.httpPool.Return(cl)
			this.GetContext().Infof("Client returned")
		} else {
			this.GetContext().Infof("Client Borrowed was nil")
		}
		this.wg.Done()

	}()

	if cl != nil {
		client := cl.(*http.Client)

		this.GetContext().Infof("Client Borrowed")

		this.GetContext().Infof("Making Request")
		this.MakeRequest(client)
		this.GetContext().Infof("Request completed normally")
	} else {
		this.GetContext().Infof("Failed to borrow a http client, received nil instead")
	}

}

func (this *SearchStage) Run() {
	this.HandleIn()
}
