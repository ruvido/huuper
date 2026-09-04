package admin

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Handlers struct {
	Settings              func(e *core.RequestEvent) error
	Summary               func(e *core.RequestEvent) error
	GroupsList            func(e *core.RequestEvent) error
	GroupGet              func(e *core.RequestEvent) error
	GroupAssistant        func(e *core.RequestEvent) error
	EventsList            func(e *core.RequestEvent) error
	EventCreate           func(e *core.RequestEvent) error
	EventUpdate           func(e *core.RequestEvent) error
	EventReschedule       func(e *core.RequestEvent) error
	EventCancel           func(e *core.RequestEvent) error
	EventCancelOccurrence func(e *core.RequestEvent) error
	EventSetActive        func(e *core.RequestEvent) error
	EventAttendance       func(e *core.RequestEvent) error
	RequestsList          func(e *core.RequestEvent) error
	RequestCreate         func(e *core.RequestEvent) error
	RequestGet            func(e *core.RequestEvent) error
	RequestAction         func(e *core.RequestEvent) error
	EventDetails          func(e *core.RequestEvent) error
	UserGet               func(e *core.RequestEvent) error
	EventEmail            func(e *core.RequestEvent) error
	RegistrationApprove   func(e *core.RequestEvent) error
	RegistrationReject    func(e *core.RequestEvent) error
	RegistrationCancel    func(e *core.RequestEvent) error
	GroupSyncMemberships  func(e *core.RequestEvent) error
	UsersList             func(e *core.RequestEvent) error
	UserDelete            func(e *core.RequestEvent) error
	UserCancel            func(e *core.RequestEvent) error

	RetreatsList               func(e *core.RequestEvent) error
	RetreatCreate              func(e *core.RequestEvent) error
	RetreatUpdate              func(e *core.RequestEvent) error
	RetreatCancel              func(e *core.RequestEvent) error
	RetreatDetails             func(e *core.RequestEvent) error
	RetreatUploadGallery       func(e *core.RequestEvent) error
	RetreatRemoveGalleryFile   func(e *core.RequestEvent) error
	RetreatBroadcastEmail      func(e *core.RequestEvent) error
	RetreatRegistrationApprove func(e *core.RequestEvent) error
	RetreatRegistrationReject  func(e *core.RequestEvent) error
	RetreatRegistrationCancel  func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/admin/settings/{name}", backendinternal.AdminOnly(h.Settings)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/summary", backendinternal.AdminOnly(h.Summary)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/groups", backendinternal.AdminOnly(h.GroupsList)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/groups/{id}", backendinternal.AdminOnly(h.GroupGet)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/{id}/assistant", backendinternal.AdminOnly(h.GroupAssistant)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events", backendinternal.AdminOnly(h.EventsList)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events", backendinternal.AdminOnly(h.EventCreate)).Bind(apis.RequireAuth())
	se.Router.PATCH("/api/admin/events/{id}", backendinternal.AdminOnly(h.EventUpdate)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/reschedule", backendinternal.AdminOnly(h.EventReschedule)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/cancel", backendinternal.AdminOnly(h.EventCancel)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/cancel-occurrence", backendinternal.AdminOnly(h.EventCancelOccurrence)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/active", backendinternal.AdminOnly(h.EventSetActive)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/attendance", backendinternal.AdminOnly(h.EventAttendance)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/requests", backendinternal.AdminOnly(h.RequestsList)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/requests", backendinternal.AdminOnly(h.RequestCreate)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/requests/{id}", backendinternal.AdminOnly(h.RequestGet)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/requests/{id}/actions", backendinternal.AdminOnly(h.RequestAction)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events/{id}", backendinternal.AdminOnly(h.EventDetails)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/email", backendinternal.AdminOnly(h.EventEmail)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/approve", backendinternal.AdminOnly(h.RegistrationApprove)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/reject", backendinternal.AdminOnly(h.RegistrationReject)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/cancel", backendinternal.AdminOnly(h.RegistrationCancel)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/sync-memberships", backendinternal.AdminOnly(h.GroupSyncMemberships)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/users", backendinternal.AdminOnly(h.UsersList)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/users/{id}", backendinternal.AdminOnly(h.UserGet)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/users/{id}/cancel", backendinternal.AdminOnly(h.UserCancel)).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/users/{id}", backendinternal.AdminOnly(h.UserDelete)).Bind(apis.RequireAuth())

	se.Router.GET("/api/admin/retreats", backendinternal.AdminOnly(h.RetreatsList)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreats", backendinternal.AdminOnly(h.RetreatCreate)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/retreats/{id}", backendinternal.AdminOnly(h.RetreatDetails)).Bind(apis.RequireAuth())
	se.Router.PATCH("/api/admin/retreats/{id}", backendinternal.AdminOnly(h.RetreatUpdate)).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/retreats/{id}", backendinternal.AdminOnly(h.RetreatCancel)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreats/{id}/gallery", backendinternal.AdminOnly(h.RetreatUploadGallery)).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/admin/retreats/{id}/gallery/{filename}", backendinternal.AdminOnly(h.RetreatRemoveGalleryFile)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreats/{id}/broadcast-email", backendinternal.AdminOnly(h.RetreatBroadcastEmail)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreat-registrations/{id}/approve", backendinternal.AdminOnly(h.RetreatRegistrationApprove)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreat-registrations/{id}/reject", backendinternal.AdminOnly(h.RetreatRegistrationReject)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/retreat-registrations/{id}/cancel", backendinternal.AdminOnly(h.RetreatRegistrationCancel)).Bind(apis.RequireAuth())
}
