package web

import (
	"html/template"
	"io"
)

var RootPageTemplate = template.Must(template.ParseGlob("pages/common/*"))

type Page struct {
	Content template.HTML
	Title   string
	Heading string
	CSS     template.HTML
	JS      template.HTML
}

func (p Page) Render(w io.Writer) error {
	// Render the template to the HTTP response.
	return RootPageTemplate.ExecuteTemplate(w, "page", p)
}
