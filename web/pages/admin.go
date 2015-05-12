package web

import (
	"appengine"
	"html/template"
	"net/http"
)

func init() {
	http.HandleFunc("/admin", handler)
}

type AdminPage struct {
	Page
}

func NewAdminPage() *AdminPage {
	admin := AdminPage{}
	admin.Title = "Admin Page"
	admin.Heading = "Admin Page"

	return &admin
}

func handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := NewAdminPage()
	p.Content = ""
	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		c.Errorf("Rendering templabbte: %v", err)
	}
}
