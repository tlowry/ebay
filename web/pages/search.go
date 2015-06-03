package web

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"github.com/tlowry/ebay/pipeline"
	"html/template"
	"net/http"
)

var itemCell *template.Template
var itemCellErr error
var actionPanel = `<div id="actionPanel">
<button id="reportSelected">A button element</button>
</div>`

func init() {
	http.HandleFunc("/search", search)

	itemCell, itemCellErr = template.New("itemCell").Parse(`
	<tr id="{{.ListingId}}" class="itemRow">
		<td class="tdCheck"><input type="checkbox" name="selectItemBox" class="selectItemBox"></input></td>
		<td class="tdTier">{{.Tier}}</td>
		<td class="tdCurrency">{{.Currency}}</td>
		<td class="tdPrice"> {{.Price}}</td>
		<td class="tdDesc">{{.Description}}</td>
		<td class="tdTime">{{.FormatTime}}</td>
		<td class="tdURL"><a href="{{.Url }}"><img width="225" height="225" src="{{.ImageUrl}}"></a></td>
	</tr>
	`)
}

func renderItem(w http.ResponseWriter, item pipeline.EbayItem) {
	itemCell.ExecuteTemplate(w, "itemCell", item)
}

func search(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	p := Page{}
	p.Title = "Search"
	p.Heading = "Search Page"

	//Render the table
	if itemCellErr == nil {
		ctx.Infof("Serving search request normally")
	} else {
		ctx.Infof("Serving search request, %s", itemCellErr.Error())
	}

	p.JS = template.HTML(`
	<script type="text/javascript" src="files/js/jquery.tablesorter.js"></script>
	<script src="files/js/search.js"></script>
	`)

	q := datastore.NewQuery("EbayItem")
	it := q.Run(ctx)

	var item pipeline.EbayItem
	var err error

	count := 0

	text := bytes.NewBufferString(`<table id="resultsTable" class="tablesorter"><thead class="ui-widget-header"><th>Select</th><th>Tier</th><th>Currency</th><th>Price</th><th>Description</th><th>Expiry</th><th>Image</th></thead><tbody class="ui-widget-content">`)

	for _, err = it.Next(&item); err == nil; _, err = it.Next(&item) {

		err = itemCell.Execute(text, item)

		if err != nil {
			ctx.Errorf("Hit error rendering %s", err.Error())
		}

		count++
	}

	if err == datastore.Done {
		ctx.Infof("Query completed normally and rendered %d items", count)
	} else {
		ctx.Errorf("Query ended in error %s after %d items", err.Error(), count)
	}

	text.WriteString(`</tbody></table>`)
	text.WriteString(actionPanel)

	dialog := Modal{}
	dialog.Title = "Hello"
	dialog.Content = "There"
	dialog.Id = "modalDial"
	dialog.Render(text)

	// DONE

	p.Content = template.HTML(text.String())

	// Render the template to the HTTP response.
	if err := p.Render(w); err != nil {
		ctx.Errorf("Error Rendering Page: %v", err)
	}
}
