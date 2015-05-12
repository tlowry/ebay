package stages

import (
	"appengine"
	"appengine/datastore"
	"github.com/jbrukh/bayesian"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"log"
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

func (fs *FilterStage) HandleIn() {

	const (
		Wanted   bayesian.Class = "Wanted"
		UnWanted bayesian.Class = "UnWanted"
	)
	defer close(fs.Out)
	classifier, err := bayesian.NewClassifierFromFile("files/gfx.ebay.classifier")
	if err == nil {
		for item := range fs.In {
			log.Println("filter recvd ", item)

			q := datastore.NewQuery("EbayItem").Filter("ListingId =", item.ListingId).Limit(1)
			res := q.Run(fs.GetContext())

			var oldItem *pipeline.EbayItem
			res.Next(oldItem)

			if oldItem != nil {
				descWords := strings.Split(item.Description, " ")

				_, inx, _ := classifier.ProbScores(descWords)
				class := classifier.Classes[inx]

				switch class {
				case Wanted:
					log.Println("FilterStage: ", item.Description, " is wanted")
					fs.Out <- item
				case UnWanted:
					log.Println("FilterStage: ", item.Description, " is unwanted")
				default:
					log.Println("FilterStage: ", item.Description, " is unknown class: ", class)
				}

			} else {
				log.Println("FilterStage: ", item.Description, "  already in db")
			}

		}
	} else {
		log.Println("Failed to open classifer ", err)
	}

}

func (fs *FilterStage) Run() {
	fs.HandleIn()
}
