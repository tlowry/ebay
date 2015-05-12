package web

import (
	"net/http"
)

func init() {
	fs := http.FileServer(http.Dir("files"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))
}
