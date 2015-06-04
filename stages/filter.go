package stages

import (
	"appengine"
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

func (fs *FilterStage) HandleIn() {

	const (
		GFX      bayesian.Class = "gfx"
		CPU      bayesian.Class = "cpu"
		APU      bayesian.Class = "apu"
		System   bayesian.Class = "system"
		Unwanted bayesian.Class = "unwanted"
	)

	defer close(fs.Out)

	ctx := fs.GetContext()
	classifier, err := bayesian.NewClassifierFromFile("conf/item.ebay.classifier")

	if err == nil {
		seen := make(map[string]bool)
		count := 0
		dupeCount := 0
		wantedCount := 0
		for item := range fs.In {
			count++

			if seen[item.ListingId] {
				dupeCount++
			} else {
				seen[item.ListingId] = true
				descWords := strings.Split(item.Description, " ")

				_, inx, _ := classifier.ProbScores(descWords)
				class := classifier.Classes[inx]

				switch class {
				case GFX:
					fs.Out <- item
					wantedCount++
				case CPU:
					ctx.Infof("FilterStage: ", item.Description, " is a cpu")
				case APU:
					ctx.Infof("FilterStage: ", item.Description, " is an apu")
				case System:
					ctx.Infof("FilterStage: ", item.Description, " is a system")
				case Unwanted:
					//ctx.Infof("FilterStage: ", item.Description, " is unwanted")
				default:
					ctx.Infof("FilterStage: ", item.Description, " is unknown class: ", class)
				}
			}

		}
		ctx.Infof("Filter recv %d total %d dupe and %d wanted items", count, dupeCount, wantedCount)
	} else {
		ctx.Infof("Failed to open classifer ", err)
	}

}

func (fs *FilterStage) Run() {
	fs.HandleIn()
}
