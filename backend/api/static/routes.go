package static

import (
	"os"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func Register(se *core.ServeEvent) {
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
	se.Router.GET("/retreat/{slug}", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat/index.html")
	})
	se.Router.GET("/retreat/{slug}/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat/index.html")
	})
	se.Router.GET("/retreat-payment/", func(e *core.RequestEvent) error {
		return e.FileFS(site, "retreat-payment/index.html")
	})

	se.Router.GET("/{path...}", apis.Static(site, false))
}
