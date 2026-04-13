package me

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type Handlers struct {
	Settings           func(e *core.RequestEvent) error
	GroupsList         func(e *core.RequestEvent) error
	GroupGet           func(e *core.RequestEvent) error
	GroupAssistant     func(e *core.RequestEvent) error
	UserGet            func(e *core.RequestEvent) error
	EventGet           func(e *core.RequestEvent) error
	EventStatus        func(e *core.RequestEvent) error
	EventUnsubscribe   func(e *core.RequestEvent) error
	TelegramToken      func(e *core.RequestEvent) error
	RequestsList       func(e *core.RequestEvent) error
	RequestGet         func(e *core.RequestEvent) error
	RequestAction      func(e *core.RequestEvent) error
	GroupMembers       func(e *core.RequestEvent) error
	GroupGuardians     func(e *core.RequestEvent) error
	GroupRequestsCount func(e *core.RequestEvent) error
	DefaultInvite      func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/me/settings/{name}", h.Settings).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups", h.GroupsList).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}", h.GroupGet).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/groups/{id}/assistant", h.GroupAssistant).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/users/{id}", h.UserGet).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/events/{id}", h.EventGet).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/events/{slug}/status", h.EventStatus).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{slug}/unsubscribe", h.EventUnsubscribe).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/telegram/token", h.TelegramToken).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/requests", h.RequestsList).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/requests/{id}", h.RequestGet).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/requests/{id}/actions", h.RequestAction).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/members", h.GroupMembers).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/guardians", h.GroupGuardians).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/requests-count", h.GroupRequestsCount).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/default/invite", h.DefaultInvite).Bind(apis.RequireAuth())
}
