package admin

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Handlers struct {
	Settings             func(e *core.RequestEvent) error
	Summary              func(e *core.RequestEvent) error
	GroupsList           func(e *core.RequestEvent) error
	GroupGet             func(e *core.RequestEvent) error
	GroupAssistant       func(e *core.RequestEvent) error
	EventsList           func(e *core.RequestEvent) error
	RequestsList         func(e *core.RequestEvent) error
	RequestGet           func(e *core.RequestEvent) error
	RequestAction        func(e *core.RequestEvent) error
	EventDetails         func(e *core.RequestEvent) error
	UserGet              func(e *core.RequestEvent) error
	EventEmail           func(e *core.RequestEvent) error
	RegistrationApprove  func(e *core.RequestEvent) error
	RegistrationReject   func(e *core.RequestEvent) error
	RegistrationCancel   func(e *core.RequestEvent) error
	GroupSyncMemberships func(e *core.RequestEvent) error
	UserDelete           func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/admin/settings/{name}", backendinternal.AdminOnly(h.Settings)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/summary", backendinternal.AdminOnly(h.Summary)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/groups", backendinternal.AdminOnly(h.GroupsList)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/groups/{id}", backendinternal.AdminOnly(h.GroupGet)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/{id}/assistant", backendinternal.AdminOnly(h.GroupAssistant)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events", backendinternal.AdminOnly(h.EventsList)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/requests", backendinternal.AdminOnly(h.RequestsList)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/requests/{id}", backendinternal.AdminOnly(h.RequestGet)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/requests/{id}/actions", backendinternal.AdminOnly(h.RequestAction)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events/{id}", backendinternal.AdminOnly(h.EventDetails)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/email", backendinternal.AdminOnly(h.EventEmail)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/approve", backendinternal.AdminOnly(h.RegistrationApprove)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/reject", backendinternal.AdminOnly(h.RegistrationReject)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/cancel", backendinternal.AdminOnly(h.RegistrationCancel)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/sync-memberships", backendinternal.AdminOnly(h.GroupSyncMemberships)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/users/{id}", backendinternal.AdminOnly(h.UserGet)).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/users/{id}", backendinternal.AdminOnly(h.UserDelete)).Bind(apis.RequireAuth())
}
