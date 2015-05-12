package web

import (
	"appengine"
	"html/template"
	"net/http"
)

func init() {
	http.HandleFunc("/", home)
}

func home(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Home"
	p.Heading = "Welcome to Ebayscraper Home"

	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		c.Errorf("Rendering templabbte: %v", err)
	}
}
