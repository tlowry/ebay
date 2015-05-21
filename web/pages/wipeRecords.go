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
	amt := 0
	ctx.Infof("About to delete %d items", len(keys))
	err := datastore.DeleteMulti(ctx, keys)
	if err == nil {
		amt = len(keys)
		ctx.Infof("Succesfully Deleted %d items", amt)
	} else {
		ctx.Infof("failed to delete %d items %s", amt, err.Error())
	}

	return amt
}

func wipeTask(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	ctx.Infof("Performing wipe task")
	q := datastore.NewQuery("EbayItem").KeysOnly()

	ctx.Infof("built query")
	count := 0

	buf := make([]*datastore.Key, 0, 500)

	it := q.Run(ctx)
	var key *datastore.Key
	var err error

	running := true

	for key, err = it.Next(key); running; key, err = it.Next(key) {

		if err == nil {
			if key == nil {
				ctx.Infof("Nil key")
			} else if len(buf) < cap(buf) {
				buf = append(buf, key)
			} else {
				count = deleteAll(ctx, buf)
				ctx.Infof("Deleted batch of %d entries", count)
				buf = make([]*datastore.Key, 0, 500)
			}

		} else {
			running = false
			if err == datastore.Done {
				ctx.Infof("Completed normally ", err.Error())
			} else {
				ctx.Infof("Feck ", err.Error())
			}
		}

	}

	if len(buf) > 0 {
		//count = deleteAll(ctx, buf)
		err = datastore.DeleteMulti(ctx, buf)

		if err != nil {
			ctx.Infof("Failed to delete %", err.Error())
		} else {
			ctx.Infof("Deleted remaining %d entries", count)
		}
	} else {
		ctx.Infof("Have no entries to delete")
	}

	if err == datastore.Done {
		ctx.Infof("DB wipe completed normally")
	} else if err == nil {
		ctx.Infof("DB wipe completed successfully")
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
