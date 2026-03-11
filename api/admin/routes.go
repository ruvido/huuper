package admin

import (
	backendinternal "members/internal"

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
	se.Router.GET("/api/admin/settings/{name}", backendinternal.AdminOnly(h.Settings)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/summary", backendinternal.AdminOnly(h.Summary)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events/{id}", backendinternal.AdminOnly(h.EventDetails)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/email", backendinternal.AdminOnly(h.EventEmail)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/approve", backendinternal.AdminOnly(h.RegistrationApprove)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/reject", backendinternal.AdminOnly(h.RegistrationReject)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/cancel", backendinternal.AdminOnly(h.RegistrationCancel)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/sync-memberships", backendinternal.AdminOnly(h.GroupSyncMemberships)).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/users/{id}", backendinternal.AdminOnly(h.UserDelete)).Bind(apis.RequireAuth())
}
