package web

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"github.com/tlowry/ebay/pipeline"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func init() {
	http.HandleFunc("/search", search)
}

func search(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Search"
	p.Heading = "Search Page"

	//Render the table
	ctx.Infof("Serving search request")
	p.JS = template.HTML(`
	<script type="text/javascript" src="files/js/jquery.tablesorter.js"></script>
	<script src="files/js/search.js"></script>
	`)

	text := "<table id=\"resultsTable\" class=\"tablesorter\"><thead class=\"ui-widget-header\"><th>Tier</th><th>Currency</th><th>Price</th><th>Description</th><th>Expiry</th><th>Image</th></thead><tbody class=\"ui-widget-content\">"
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

	text += "</tbody></table>"

	// DONE

	p.Content = template.HTML(text)

	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		ctx.Errorf("Rendering templabbte: %v", err)
	}
}

var tableTemplate = `
	<table id="resultsTable" class="tablesorter"><thead class="ui-widget-header"><th>Tier</th><th>Currency</th><th>Price</th><th>Description</th><th>Expiry</th><th>Image</th></thead><tbody class="ui-widget-content">

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
		<tr id="" + item.ListingId + "\">"
		text += "<td>" + item.Tier + "</td>"
		text += "<td>" + item.Currency + "</td>"
		text += "<td>" + strconv.FormatFloat(item.Price, 'f', 2, 64) + "</td>"
		text += "<td>" + item.Description + "</td>"
		text += "<td>" + seconds + "</td>"
		text += "<td><a href=\"" + item.Url + "\"><img src=\"" + item.ImageUrl + "\"></a></td>"
		text += "</tr>"
		count++
	}

	</tbody></table>
`
