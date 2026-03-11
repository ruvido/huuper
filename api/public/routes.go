package public

import "github.com/pocketbase/pocketbase/core"

type Handlers struct {
	Settings       func(e *core.RequestEvent) error
	EventsAccept   func(e *core.RequestEvent) error
	EventsRegister func(e *core.RequestEvent) error
	RequestsCreate func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/public/settings/{name}", h.Settings)
	se.Router.GET("/api/public/events/accept", h.EventsAccept)
	se.Router.POST("/api/public/events/{slug}/register", h.EventsRegister)
	se.Router.POST("/api/public/requests", h.RequestsCreate)
}
