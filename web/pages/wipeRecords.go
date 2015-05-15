package web

import (
	"appengine"
	"appengine/datastore"
	"html/template"
	"net/http"
)

func init() {
	http.HandleFunc("/task/items/wipe", wipeTask)
}

func deleteAll(ctx appengine.Context, keys []*datastore.Key) int {
	amt := len(keys)
	ctx.Infof("About to delete %d items", amt)
	err := datastore.DeleteMulti(ctx, keys)
	if err == nil {
		ctx.Infof("Succesfully Deleted %d items", amt)
	} else {
		ctx.Infof("failed to delete %d items %s", amt, err.Error())
		amt = 0
	}

	return amt
}

func wipeTask(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx.Infof("Performing wipe task")
	q := datastore.NewQuery("EbayItem").KeysOnly()

	ctx.Infof("built query")
	count := 0

	buf := make([]*datastore.Key, 0, 5000)

	it := q.Run(ctx)
	var key datastore.Key
	var err error
	for _, err = it.Next(key); err == nil; _, err = it.Next(key) {

		if len(buf) < cap(buf) {
			buf = append(buf, &key)
		} else {
			count = deleteAll(ctx, buf)
		}
	}

	if len(buf) > 0 {
		count = deleteAll(ctx, buf)
	}

	if err == datastore.Done {
		ctx.Infof("DB wipe completed normally")
	} else {
		ctx.Infof("DB wipe failed with error %s", err.Error())
	}

	ctx.Infof("Deleted %d keys", count)

	var tmpl = template.Must(template.ParseGlob("pages/common/*"))

	p := Page{}
	p.Title = "Admin Page"
	p.Content = `
		Wipe operation complete
	`
	// Render the template to the HTTP response.
	if err := tmpl.ExecuteTemplate(w, "page", p); err != nil {
		ctx.Errorf("Rendering template: %v", err)
	}
}
