package web

import (
	"appengine"
	"appengine/taskqueue"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
)

func init() {
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/wipe", wipeHandler)
	http.HandleFunc("/admin/refresh", refreshHandler)
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
	<a href="/admin/refresh">Reload DB</a>
	<a href="https://console.developers.google.com">Console</a>
	`
	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		c.Errorf("Rendering template: %v", err)
	}
}

func wipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx.Infof("Serving Wipe request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := taskqueue.NewPOSTTask("/task/items/wipe", url.Values{})
	_, err := taskqueue.Add(ctx, t, "")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Admin Page"

	if err != nil {
		p.Content = template.HTML(fmt.Sprintf("Error scheduling wipe task", err.Error()))

	} else {
		p.Content = template.HTML(fmt.Sprintf("Succesfully scheduled wipe task"))
	}

	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		ctx.Errorf("Rendering templabbte: %v", err)
	}
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx.Infof("Serving refresh request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := taskqueue.NewPOSTTask("/task/items/refresh", url.Values{})
	_, err := taskqueue.Add(ctx, t, "")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Admin Page"

	if err != nil {
		p.Content = template.HTML(fmt.Sprintf("Error scheduling wipe task", err.Error()))

	} else {
		p.Content = template.HTML(fmt.Sprintf("Succesfully scheduled wipe task"))
	}

	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		ctx.Errorf("Rendering template: %v", err)
	}
}
