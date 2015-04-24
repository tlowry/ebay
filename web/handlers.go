package web

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	"github.com/tlowry/ebay/app"
	"github.com/tlowry/ebay/pipeline"
	"github.com/tlowry/ebay/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func refreshTask(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	ctx.Infof("About to parse args")
	var confFile = "files/hierarchy.json"

	fileBytes, err := ioutil.ReadFile(confFile)
	util.FailOnErr("failed to read json conf", err)

	conf := util.TierConf{}
	err = json.Unmarshal(fileBytes, &conf)
	util.FailOnErr("Failed to parse json", err)

	ctx.Infof("About to run tiers ", conf.MaxPrice)

	if len(conf.Tiers) < 1 {
		util.Fail("ERROR: no tiers in json")
	}

	searcher := app.NewTierSearch(&conf, ctx)
	searcher.RunSearch()

	fmt.Fprint(w, "Refresh task complete")
}

func searchTask(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)
	ctx.Infof("Serving search request")

	text := "<html><head>"

	// scripts
	text += `
	<script src="../files/js/jquery-2.1.3.min.js"></script>
	<script type="text/javascript" src="../files/js/jquery.tablesorter.js"></script>
	<script src="../files/js/main.js"></script>
	<script src="../files/js/jquery-ui-1.11.3/jquery-ui.min.js"></script>
	<link rel="stylesheet" type="text/css" href="../files/js/jquery-ui-1.11.3/jquery-ui.css">
	`

	text += "</head><body><table id=\"resultsTable\" class=\"tablesorter\"><thead class=\"ui-widget-header\"><th>Tier</th><th>Currency</th><th>Price</th><th>Description</th><th>Expiry</th><th>Image</th></thead><tbody class=\"ui-widget-content\">"

	now := time.Now()

	q := datastore.NewQuery("EbayItem")
	var items []pipeline.EbayItem
	_, err := q.GetAll(ctx, &items)

	if err != nil {
		ctx.Infof("Error in query: ", err.Error())
	}

	count := 0

	for _, item := range items {

		ctx.Infof("Rendering ", item)
		seconds := ""
		if !item.ExpiryDate.IsZero() {
			seconds = fmt.Sprintf("%f", item.ExpiryDate.Sub(now).Seconds())
		}
		text += "<tr id=\"" + item.ListingId + "\">"
		text += "<td>" + item.Tier + "</td>"
		text += "<td>" + item.Currency + "</td>"
		text += "<td>" + strconv.FormatFloat(item.Price, 'f', 2, 64) + "</td>"
		text += "<td>" + item.Description + "</td>"
		text += "<td>" + seconds + "</td>"
		text += "<td><a href=\"" + item.Url + "\"><img src=\"" + item.ImageUrl + "\"></a></td>"
		text += "</tr>"
		count++
	}

	text += "</tbody></table></body></html>"
	fmt.Fprint(w, text)
}

func init() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	http.HandleFunc("/tasks/refreshmerchants", refreshTask)

	http.HandleFunc("/tasks/search", searchTask)
}
