package web

import (
	"html/template"
	"io"
)

var ModalTemplate = template.Must(template.ParseGlob("pages/common/modal.html"))

type Modal struct {
	Content template.HTML
	Id      string
	Title   string
}

func (m Modal) Render(w io.Writer) error {
	// Render the template to the HTTP response.
	return ModalTemplate.ExecuteTemplate(w, "modal", m)
}
