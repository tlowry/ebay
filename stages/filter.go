package stages

import (
	"appengine"
	"appengine/datastore"
	"github.com/jbrukh/bayesian"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"strings"
)

type FilterStage struct {
	*pipeline.Stage
	conf *util.TierConf
}

func NewFilterStage(conf *util.TierConf, ctx appengine.Context) *FilterStage {
	filter := FilterStage{}
	filter.Stage = pipeline.NewStage(ctx)
	filter.conf = conf
	return &filter
}

func (fs *FilterStage) Init() {
	fs.SetName("FilterStage")
	fs.Stage.Init()
}

func (fs FilterStage) checkExists(item *pipeline.EbayItem) bool {

	fs.GetContext().Infof("Starting check")
	q := datastore.NewQuery("EbayItem").Filter("ListingId =", item.ListingId).Limit(1)
	fs.GetContext().Infof("Query ready")
	res := q.Run(fs.GetContext())
	fs.GetContext().Infof("Query complete")

	var oldItem *pipeline.EbayItem = nil
	res.Next(oldItem)
	fs.GetContext().Infof("result acquired ", oldItem)

	return oldItem != nil

}

func (fs *FilterStage) HandleIn() {

	const (
		Wanted   bayesian.Class = "Wanted"
		UnWanted bayesian.Class = "UnWanted"
	)
	defer close(fs.Out)

	ctx := fs.GetContext()
	classifier, err := bayesian.NewClassifierFromFile("conf/gfx.ebay.classifier")
	if err == nil {
		for item := range fs.In {
			ctx.Infof("filter recvd ", item)

			if true || fs.checkExists(&item) {
				descWords := strings.Split(item.Description, " ")

				_, inx, _ := classifier.ProbScores(descWords)
				class := classifier.Classes[inx]

				switch class {
				case Wanted:
					ctx.Infof("FilterStage: ", item.Description, " is wanted")
					fs.Out <- item
				case UnWanted:
					ctx.Infof("FilterStage: ", item.Description, " is unwanted")
				default:
					ctx.Infof("FilterStage: ", item.Description, " is unknown class: ", class)
				}

			} else {
				ctx.Infof("FilterStage: ", item.Description, "  already in db")
			}

		}
	} else {
		ctx.Infof("Failed to open classifer ", err)
	}

}

func (fs *FilterStage) Run() {
	fs.HandleIn()
}
