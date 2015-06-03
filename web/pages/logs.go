// This sample gets the app displays 5 log Records at a time, including all
// AppLogs, with a Next link to let the user page through the results using the
// Record's Offset property.
package web

import (
	"encoding/base64"
	"html/template"
	"net/http"

	"appengine"
	"appengine/log"
	"bytes"
	"golang.org/x/text/encoding"
)

func init() {
	http.HandleFunc("/admin/logs", logsHandler)
}

const recordsPerPage = 15

func logsHandler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Set up a data structure to pass to the HTML template.
	var data struct {
		Records []*log.Record
		Offset  string // base-64 encoded string
	}

	// Set up a log.Query.
	query := &log.Query{AppLogs: true}

	// Get the incoming offset param from the Next link to advance through
	// the logs. (The first time the page is loaded there won't be any offset.)
	if offset := r.FormValue("offset"); offset != "" {
		query.Offset, _ = base64.URLEncoding.DecodeString(offset)
	}

	// Run the query, obtaining a Result iterator.
	res := query.Run(c)

	// Iterate through the results populating the data struct.
	for i := 0; i < recordsPerPage; i++ {
		rec, err := res.Next()
		if err == log.Done {
			break
		}
		if err != nil {
			c.Errorf("Reading log records: %v", err)
			break
		}

		count := 0
		for _, msg := range rec.AppLogs {

			buf := bytes.NewBufferString("")
			enc := encoding.Encoding.NewEncoder(encoding.Replacement)
			enc.Transform(buf.Bytes(), []byte(msg.Message), false)
			msg.Message = buf.String()
			count++
		}
		if count > 0 {
			data.Records = append(data.Records, rec)
		}

		if i == recordsPerPage-1 {
			data.Offset = base64.URLEncoding.EncodeToString(rec.Offset)
		}
	}

	var b []byte
	buf := bytes.NewBuffer(b)

	// Render the template to the HTTP response.
	if err := tmpl.Execute(buf, data); err != nil {
		c.Errorf("Rendering template: %v", err)
	}

	page := Page{}
	page.Content = template.HTML(string(buf.Bytes()))

	err := page.Render(w)
	if err != nil {
		c.Infof("Error Rendering Page %s", err.Error())
	}

}

var tmpl = template.Must(template.New("").Parse(`
        {{range .Records}}
                <h2>Request Log</h2>
                <p>{{.EndTime}}: {{.IP}} {{.Method}} {{.Resource}}</p>
                {{with .AppLogs}}
                        <h3>App Logs:</h3>
                        <ul>
                        {{range .}}
                                <li><pre>{{.Time}}: {{.Message}}</pre></li>
                        <{{end}}
                        </ul>
                {{end}}
        {{end}}
        {{with .Offset}}
                <a href="?offset={{.}}">Next</a>
        {{end}}
`))
