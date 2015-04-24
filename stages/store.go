package stages

import (
	"appengine"
	"appengine/datastore"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
)

type DataStoreStage struct {
	*pipeline.Stage
	conf    *util.TierConf
	context appengine.Context
}

func NewDataStoreStage(ctx appengine.Context) *DataStoreStage {
	ps := DataStoreStage{}
	ps.Stage = pipeline.NewStage(ctx)
	ps.context = ctx
	return &ps
}

func (this *DataStoreStage) Init() {
	this.SetName("DataStoreStage")
	this.Stage.Init()
}

func ItemKey(c appengine.Context) *datastore.Key {
	// The string "default_ebay_items" here could be varied to have multiple items.
	return datastore.NewKey(c, "EbayItems", "default_ebay_items", 0, nil)
}

func (this *DataStoreStage) HandleIn() {

	ctx := this.GetContext()
	for item := range this.In {

		ctx.Infof("Persisting ", item.Description)

		key := datastore.NewIncompleteKey(this.context, "EbayItem", ItemKey(this.context))
		_, err := datastore.Put(this.context, key, &item)

		if err != nil {
			ctx.Infof("Persisting ", string(err.Error()))
		}
	}

}

func (this *DataStoreStage) Run() {
	this.HandleIn()
}
