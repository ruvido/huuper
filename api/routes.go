package api

import (
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterRoutes(app *pocketbase.PocketBase, se *core.ServeEvent) {
	se.Router.GET("/api/settings/{name}", GetSettingsHandler(app))
	se.Router.GET("/api/events/accept", AcceptEventHandler(app))
	se.Router.POST("/api/events/{slug}/register", RegisterEventHandler(app))

	se.Router.GET("/api/admin/summary", AdminSummaryHandler(app)).Bind(apis.RequireAuth())
	se.Router.GET("/api/admin/events/{id}", AdminEventDetailsHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/events/{id}/email", AdminEventEmailHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/approve", AdminApproveRegistrationHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/reject", AdminRejectRegistrationHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/registrations/{id}/cancel", AdminCancelRegistrationHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/admin/groups/sync-memberships", AdminSyncGroupMembershipsHandler(app)).Bind(apis.RequireAuth())

	se.Router.GET("/api/events/{slug}/status", EventStatusHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/events/{slug}/unsubscribe", EventUnsubscribeHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/telegram/generate-token", GenerateTelegramTokenHandler(app)).Bind(apis.RequireAuth())

	se.Router.POST("/api/requests/submit", SubmitRequestHandler(app))
	se.Router.GET("/api/requests", ListRequestsHandler(app)).Bind(apis.RequireAuth())
	se.Router.GET("/api/requests/{id}", GetRequestHandler(app)).Bind(apis.RequireAuth())
	se.Router.POST("/api/requests/{id}/action", RequestActionHandler(app)).Bind(apis.RequireAuth())

	se.Router.GET("/api/groups/{id}/members", GroupMembersHandler(app)).Bind(apis.RequireAuth())
	se.Router.GET("/api/groups/{id}/guardians", GroupGuardiansHandler(app)).Bind(apis.RequireAuth())
	se.Router.GET("/api/groups/{id}/requests-count", GroupRequestsCountHandler(app)).Bind(apis.RequireAuth())
	se.Router.GET("/api/groups/default-invite", DefaultGroupInviteHandler(app)).Bind(apis.RequireAuth())

	se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
}
