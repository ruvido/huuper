package static

import (
	"io/fs"
	"os"

	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func Register(app *pocketbase.PocketBase, se *core.ServeEvent) {
	site := os.DirFS("./frontend/site")
	se.Router.GET("/admin/battleplan/view/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/view/index.html")
	})
	se.Router.GET("/admin/battleplan/view/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/view/index.html")
	})
	se.Router.GET("/admin/battleplan/edit/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/edit/index.html")
	})
	se.Router.GET("/admin/battleplan/edit/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/edit/index.html")
	})
	se.Router.GET("/me/battleplan/view/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/view/index.html")
	})
	se.Router.GET("/me/battleplan/view/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/view/index.html")
	})
	se.Router.GET("/me/battleplan/edit/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/edit/index.html")
	})
	se.Router.GET("/me/battleplan/edit/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/edit/index.html")
	})
	se.Router.GET("/admin/events/edit/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/events/edit/index.html")
	})
	se.Router.GET("/admin/events/edit/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/events/edit/index.html")
	})
	se.Router.GET("/me/events/edit/{id}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/events/edit/index.html")
	})
	se.Router.GET("/me/events/edit/{id}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/events/edit/index.html")
	})
	se.Router.GET("/admin/events/new", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/events/new/index.html")
	})
	se.Router.GET("/admin/events/new/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/events/new/index.html")
	})
	se.Router.GET("/me/events/new", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/events/new/index.html")
	})
	se.Router.GET("/me/events/new/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/events/new/index.html")
	})
	se.Router.GET("/admin/battleplan/new", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/new/index.html")
	})
	se.Router.GET("/admin/battleplan/new/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "admin/battleplan/new/index.html")
	})
	se.Router.GET("/me/battleplan/new", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/new/index.html")
	})
	se.Router.GET("/me/battleplan/new/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "me/battleplan/new/index.html")
	})
	se.Router.GET("/event/{slug}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "event/index.html")
	})
	se.Router.GET("/event/{slug}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "event/index.html")
	})
	se.Router.GET("/retreat/{slug}", retreatPage(app, site))
	se.Router.GET("/retreat/{slug}/", retreatPage(app, site))
	se.Router.GET("/retreat-payment/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat-payment/index.html")
	})
	se.Router.GET("/retreat-accept/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat-accept/index.html")
	})
	se.Router.GET("/retreat-registered/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat-registered/index.html")
	})

	se.Router.GET("/{path...}", apis.Static(site, false))
}

// retreatPage serves the public retreat page with the record's own title and
// description written into its <head>, so a link shared on WhatsApp, Telegram
// or Facebook names the retreat it points at. Every failure — unknown slug,
// unreadable file — falls back to the page exactly as the build wrote it.
func retreatPage(app *pocketbase.PocketBase, site fs.FS) func(e *core.RequestEvent) error {
	const pagePath = "retreat/index.html"

	return func(e *core.RequestEvent) error {
		retreat, err := retreatsinternal.FindBySlug(app, e.Request.PathValue("slug"))
		if err != nil || retreat == nil {
			return e.FileFS(site, pagePath)
		}

		page, err := fs.ReadFile(site, pagePath)
		if err != nil {
			return e.FileFS(site, pagePath)
		}

		e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err = e.Response.Write(retreatsinternal.RenderPageHead(page, retreat))
		return err
	}
}
