package stages

import (
	"appengine"
	"fmt"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"log"
	"os"
	"strconv"
	"time"
)

type PersistStage struct {
	*pipeline.Stage
	conf *util.TierConf
}

func NewPersistStage(ctx appengine.Context) *PersistStage {
	ps := PersistStage{}
	ps.Stage = pipeline.NewStage(ctx)
	return &ps
}

func (this *PersistStage) Init() {
	this.SetName("PersistStage")
	this.Stage.Init()
}

func (this *PersistStage) HandleIn() {
	f, err := os.Create("./output/out.html")
	if err != nil {
		return
	}
	defer f.Close()
	opener := "<html><head>"

	// scripts
	opener += `
	<script src="../files/js/jquery-2.1.3.min.js"></script>
	<script type="text/javascript" src="../files/js/jquery.tablesorter.js"></script>
	<script src="../files/js/main.js"></script>
	<script src="../files/js/jquery-ui-1.11.3/jquery-ui.min.js"></script>
	<link rel="stylesheet" type="text/css" href="../files/js/jquery-ui-1.11.3/jquery-ui.css">
	`

	opener += "</head><body><table id=\"resultsTable\" class=\"tablesorter\"><thead class=\"ui-widget-header\"><th>Tier</th><th>Currency</th><th>Price</th><th>Description</th><th>Expiry</th><th>Image</th></thead><tbody class=\"ui-widget-content\">"

	_, err = f.WriteString(opener)

	now := time.Now()

	count := 0
	for item := range this.In {

		log.Println("Persisting ", item)
		seconds := ""
		if !item.ExpiryDate.IsZero() {
			seconds = fmt.Sprintf("%f", item.ExpiryDate.Sub(now).Seconds())
		}
		str := "<tr id=\"" + item.ListingId + "\">"
		str += "<td>" + item.Tier + "</td>"
		str += "<td>" + item.Currency + "</td>"
		str += "<td>" + strconv.FormatFloat(item.Price, 'f', 2, 64) + "</td>"
		str += "<td>" + item.Description + "</td>"
		str += "<td>" + seconds + "</td>"
		str += "<td><a href=\"" + item.Url + "\"><img src=\"" + item.ImageUrl + "\"></a></td>"
		str += "</tr>"
		_, err = f.WriteString(str)
		count++
	}

	end := ``
	end += "</tbody></table></body></html>"
	_, err = f.WriteString(end)
}

func (this *PersistStage) Run() {
	this.HandleIn()
}
