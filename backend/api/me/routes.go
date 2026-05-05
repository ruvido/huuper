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
	EventsList         func(e *core.RequestEvent) error
	EventGet           func(e *core.RequestEvent) error
	EventStatus        func(e *core.RequestEvent) error
	EventCreate        func(e *core.RequestEvent) error
	EventUpdate        func(e *core.RequestEvent) error
	EventReschedule    func(e *core.RequestEvent) error
	EventCancel        func(e *core.RequestEvent) error
	EventCancelOccurrence func(e *core.RequestEvent) error
	EventRegister      func(e *core.RequestEvent) error
	EventUnregister    func(e *core.RequestEvent) error
	EventUnsubscribe   func(e *core.RequestEvent) error
	EventAttendance    func(e *core.RequestEvent) error
	TelegramToken      func(e *core.RequestEvent) error
	RequestsList       func(e *core.RequestEvent) error
	RequestGet         func(e *core.RequestEvent) error
	RequestAction      func(e *core.RequestEvent) error
	GroupMembers       func(e *core.RequestEvent) error
	GroupGuardians     func(e *core.RequestEvent) error
	GroupRequestsCount func(e *core.RequestEvent) error
	DefaultInvite      func(e *core.RequestEvent) error
	BattleplansList    func(e *core.RequestEvent) error
	BattleplanGet      func(e *core.RequestEvent) error
	BattleplanCreate   func(e *core.RequestEvent) error
	BattleplanUpdate   func(e *core.RequestEvent) error
	BattleplanStatus   func(e *core.RequestEvent) error
	BattleplanDelete   func(e *core.RequestEvent) error
	BattleplanAccess   func(e *core.RequestEvent) error
}

func Register(se *core.ServeEvent, h Handlers) {
	se.Router.GET("/api/me/settings/{name}", h.Settings).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups", h.GroupsList).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}", h.GroupGet).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/groups/{id}/assistant", h.GroupAssistant).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/users/{id}", h.UserGet).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/events", h.EventsList).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events", h.EventCreate).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/events/{id}", h.EventGet).Bind(apis.RequireAuth())
	se.Router.PATCH("/api/me/events/{id}", h.EventUpdate).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{id}/reschedule", h.EventReschedule).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{id}/cancel", h.EventCancel).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{id}/cancel-occurrence", h.EventCancelOccurrence).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{id}/register", h.EventRegister).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{id}/unregister", h.EventUnregister).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/events/{slug}/status", h.EventStatus).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/events/{slug}/unsubscribe", h.EventUnsubscribe).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/registrations/{id}/attendance", h.EventAttendance).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/telegram/token", h.TelegramToken).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/requests", h.RequestsList).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/requests/{id}", h.RequestGet).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/requests/{id}/actions", h.RequestAction).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/members", h.GroupMembers).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/guardians", h.GroupGuardians).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/{id}/requests-count", h.GroupRequestsCount).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/groups/default/invite", h.DefaultInvite).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/battleplans", h.BattleplansList).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/battleplans", h.BattleplanCreate).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/battleplans/{id}", h.BattleplanGet).Bind(apis.RequireAuth())
	se.Router.PATCH("/api/me/battleplans/{id}", h.BattleplanUpdate).Bind(apis.RequireAuth())
	se.Router.POST("/api/me/battleplans/{id}/status", h.BattleplanStatus).Bind(apis.RequireAuth())
	se.Router.DELETE("/api/me/battleplans/{id}", h.BattleplanDelete).Bind(apis.RequireAuth())
	se.Router.GET("/api/me/access/battleplan", h.BattleplanAccess).Bind(apis.RequireAuth())
}
