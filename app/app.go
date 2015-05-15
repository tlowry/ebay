package app

import (
	"appengine"
	"appengine/urlfetch"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/stages"
	"github.com/tlowry/ebay/util"
	"github.com/tlowry/grawl/browser"
	"github.com/tlowry/grawl/element"
	"strconv"
	"sync"
)

type TierSearch struct {
	conf *util.TierConf
	ctx  appengine.Context
	pipe *pipeline.Pipeline
}

func NewTierSearch(conf *util.TierConf, ctx appengine.Context) *TierSearch {
	ret := TierSearch{}
	ret.conf = conf
	ret.pipe = pipeline.NewPipeline(ctx)
	ret.ctx = ctx
	return &ret
}

func (ts *TierSearch) RunSearch() {

	cl := urlfetch.Client(ts.ctx)
	conn := browser.NewBrowserWithClient(cl)
	conn.SetUserAgent("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.137 Safari/537.36")

	ts.ctx.Infof("Loading page to parse form")
	page := conn.Load("http://www.ebay.ie/sch/ebayadvsearch/")
	ts.ctx.Infof("Page Loaded")

	form := page.ById("adv_search_from").(*element.Form)
	if form == nil {
		ts.ctx.Infof("Failed to parse form")
	} else {
		searchGroup := pipeline.NewStageGroup()
		pool := util.NewPool(8, ts.ctx)
		pool.MakeFunc = func() interface{} {
			cl := urlfetch.Client(ts.ctx)
			return cl
		}
		pool.ReturnFunc = func(item interface{}) {

		}
		var wg sync.WaitGroup
		for i, tier := range ts.conf.Tiers {
			for _, item := range tier {
				wg.Add(1)

				srch := stages.NewSearchStage(item, &wg, pool, *form, ts.ctx)
				srch.Tier = strconv.Itoa(i)
				searchGroup.AddInstance(srch)
			}
		}

		// Add the group of search stages to the pipeline
		ts.pipe.AddBack(searchGroup, 1)

		// close search>filter chan later
		ts.ctx.Infof("Creating Filter")
		filter := stages.NewFilterStage(ts.conf, ts.ctx)

		ts.ctx.Infof("Starting Filter")
		ts.pipe.AddBack(filter, 50)

		ts.ctx.Infof("Filter Started")

		ds := stages.NewDataStoreStage(ts.ctx)
		ts.pipe.AddBack(ds, 50)

		// Run
		ts.pipe.Init()
		ts.pipe.Run()

		outChan := searchGroup.GetOut()
		ts.AwaitEnd(outChan, &wg, pool)
	}

}

func (ts *TierSearch) AwaitEnd(out chan pipeline.EbayItem, wg *sync.WaitGroup, pool *util.Pool) {
	wg.Wait()
	ts.ctx.Infof("Quitting")
	close(out)
	pool.Close()
}
