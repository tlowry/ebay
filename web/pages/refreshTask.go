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
	if err == nil {
		conf := util.TierConf{}
		err = json.Unmarshal(fileBytes, &conf)
		if err == nil {
			ctx.Infof("About to run tiers ")

			if len(conf.Tiers) < 1 {
				ctx.Errorf("ERROR: no tiers in json")
			} else {
				searcher := app.NewTierSearch(&conf, ctx)

				fmt.Fprint(w, "Refresh task Running now")
				searcher.RunSearch()
			}
		} else {
			ctx.Errorf("Failed to parse json", err)
		}
	} else {
		ctx.Errorf("failed to read json conf", err)
	}

}

func init() {
	http.HandleFunc("/tasks/refreshmerchants", refreshTask)
}
