package web

import (
	"appengine"
	"appengine/datastore"
	"github.com/tlowry/ebay/pipeline"
	"html/template"
	"net/http"
)

func init() {
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/wipe", wipeHandler)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Admin Page"
	p.Content = `
	<a href="/admin/logs">Logs</a>
	<a href="/admin/wipe">Wipe DB</a>
	<a href="/tasks/refreshmerchants">Reload DB</a>
	`
	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		c.Errorf("Rendering templabbte: %v", err)
	}
}

func wipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := datastore.NewQuery("EbayItem")

	count := 0

	for t := q.Run(ctx); ; {
		var eItem pipeline.EbayItem
		key, err := t.Next(&eItem)
		if err == datastore.Done {
			ctx.Infof("Datastore query complete %s", err)
			break
		}
		if err != nil {
			ctx.Infof("Error reading item %s", err)
			break
		}

		ctx.Infof("Key=%v\nWidget=%#v\n\n", key, eItem)
		err = datastore.Delete(ctx, key)

		if err != nil {
			ctx.Infof("Error deleting %s", err.Error)
		} else {
			ctx.Infof("Deleted ok")
		}
	}

	ctx.Infof("Deleted %d keys", count)

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Admin Page"
	p.Content = `
		Wiped
	`
	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		ctx.Errorf("Rendering templabbte: %v", err)
	}
}
