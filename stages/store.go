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

func (data *DataStoreStage) Init() {
	data.SetName("DataStoreStage")
	data.Stage.Init()
}

func ItemKey(c appengine.Context) *datastore.Key {
	// The string "default_ebay_items" here could be varied to have multiple items.
	return datastore.NewKey(c, "EbayItems", "default_ebay_items", 0, nil)
}

func (data *DataStoreStage) HandleIn() {

	ctx := data.GetContext()
	for item := range data.In {

		ctx.Infof("Persisting ", item.Description)

		key := datastore.NewIncompleteKey(data.context, "EbayItem", ItemKey(data.context))
		_, err := datastore.Put(data.context, key, &item)

		if err != nil {
			ctx.Infof("Persisting ", string(err.Error()))
		}
	}

}

func (data *DataStoreStage) Run() {
	data.HandleIn()
}
