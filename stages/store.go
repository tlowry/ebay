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

func (data *DataStoreStage) writeToStore(keys *[]*datastore.Key, items *[]pipeline.EbayItem) error {

	_, err := datastore.PutMulti(data.context, *keys, *items)
	//_, err := datastore.Put(data.context, key, &item)

	*keys = make([]*datastore.Key, 0, 500)
	*items = make([]pipeline.EbayItem, 0, 500)

	return err
}
func (data *DataStoreStage) HandleIn() {

	ctx := data.GetContext()

	// Buffer batches of 500 to spare data quota
	keyBuf := make([]*datastore.Key, 0, 500)
	itemBuf := make([]pipeline.EbayItem, 0, 500)

	for item := range data.In {

		key := datastore.NewIncompleteKey(data.context, "EbayItem", ItemKey(data.context))
		keyBuf = append(keyBuf, key)
		itemBuf = append(itemBuf, item)
		ctx.Infof("Persist buffering ", item.Description)

		if len(itemBuf) == cap(itemBuf) {
			ctx.Infof("Persist buffer full, writing to datastore")
			err := data.writeToStore(&keyBuf, &itemBuf)
			if err == nil {
				ctx.Infof("Successfully wrote all buffered items to datastore")
			} else {
				ctx.Infof("Failed to write buffered items to datastore")
			}
		}

	}

	if len(keyBuf) > 0 {
		err := data.writeToStore(&keyBuf, &itemBuf)
		if err == nil {
			ctx.Infof("Successfully wrote leftover items to datastore")
		} else {
			ctx.Infof("Failed to leftover items to datastore")
		}
	} else {
		ctx.Infof("No leftover keys in buffer")
	}

}

func (data *DataStoreStage) Run() {
	data.HandleIn()
}
