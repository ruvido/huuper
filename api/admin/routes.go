package admin

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Handlers struct {
	Settings             func(e *core.RequestEvent) error
	Summary              func(e *core.RequestEvent) error
	EventDetails         func(e *core.RequestEvent) error
	EventEmail           func(e *core.RequestEvent) error
	RegistrationApprove  func(e *core.RequestEvent) error
	RegistrationReject   func(e *core.RequestEvent) error
	RegistrationCancel   func(e *core.RequestEvent) error
	GroupSyncMemberships func(e *core.RequestEvent) error
	UserDelete           func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/admin/settings/{name}", h.Settings).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/summary", h.Summary).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events/{id}", h.EventDetails).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/email", h.EventEmail).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/approve", h.RegistrationApprove).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/reject", h.RegistrationReject).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/cancel", h.RegistrationCancel).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/sync-memberships", h.GroupSyncMemberships).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/users/{id}", h.UserDelete).Bind(apis.RequireAuth())
}
