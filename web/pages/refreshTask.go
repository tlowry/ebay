package web

import (
	"appengine"
	"encoding/json"
	"fmt"
	"github.com/tlowry/ebay/app"
	"github.com/tlowry/ebay/util"
	"io/ioutil"
	"net/http"
)

func refreshTask(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	ctx.Infof("About to parse args")
	var confFile = "conf/hierarchy.json"

	fileBytes, err := ioutil.ReadFile(confFile)
	util.FailOnErr("failed to read json conf", err)

	conf := util.TierConf{}
	err = json.Unmarshal(fileBytes, &conf)
	util.FailOnErr("Failed to parse json", err)

	ctx.Infof("About to run tiers ")

	if len(conf.Tiers) < 1 {
		util.Fail("ERROR: no tiers in json")
	}

	searcher := app.NewTierSearch(&conf, ctx)

	fmt.Fprint(w, "Refresh task Running now")
	searcher.RunSearch()

}

func init() {
	http.HandleFunc("/tasks/refreshmerchants", refreshTask)
}
