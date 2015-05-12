package web

import (
	"html/template"
)

type Page struct {
	Content template.HTML
	Title   string
	Heading string
	CSS     template.HTML
	JS      template.HTML
}
