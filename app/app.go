package app

import (
	"appengine"
	"appengine/urlfetch"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/stages"
	"github.com/tlowry/ebay/util"
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
	ret.pipe = pipeline.NewPipeline()
	ret.ctx = ctx
	return &ret
}

func (ts *TierSearch) RunSearch() {
	price := ts.conf.MaxPrice
	ts.ctx.Infof("Looking for best card under %f", price)

	searchGroup := pipeline.NewStageGroup()
	pool := util.NewPool(1)
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

			srch := stages.NewSearchStage(item, &wg, pool, ts.ctx)
			srch.Tier = strconv.Itoa(i)
			searchGroup.AddInstance(srch)
		}

		if i == 2 {
			break
		}

		/*
			Todo different types of search
			use different filter criteria.

			e.g best perf for €200 =
			search from tier 1 in descending order.

			if we find an item below €200, end the search
			at that tier

		*/

	}

	// Add the group of search stages to the pipeline
	ts.pipe.AddBack(searchGroup)

	// close search>filter chan later
	ts.ctx.Infof("Creating Filter")
	filter := stages.NewFilterStage(ts.conf, ts.ctx)

	ts.ctx.Infof("Starting Filter")
	ts.pipe.AddBack(filter)

	ts.ctx.Infof("Filter Started")
	//persist := stages.NewPersistStage(ts.ctx)
	//ts.pipe.AddBack(persist)

	ds := stages.NewDataStoreStage(ts.ctx)
	ts.pipe.AddBack(ds)

	// Run
	ts.pipe.Init()
	ts.pipe.Run()

	outChan := searchGroup.GetOut()
	ts.AwaitEnd(outChan, &wg, pool)
}

func (ts *TierSearch) AwaitEnd(out chan pipeline.EbayItem, wg *sync.WaitGroup, pool *util.Pool) {
	wg.Wait()
	ts.ctx.Infof("Quitting")
	close(out)
	pool.Close()
}
