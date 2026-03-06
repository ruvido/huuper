package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"members/api"
	"members/bot"
	_ "members/migrations"
)

func init() {
	// Load .env file BEFORE migrations run
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	app := pocketbase.New()
	api.RegisterSettingsValidationHooks(app)
	api.RegisterGroupsValidationHooks(app)
	api.RegisterUsersNormalizationHooks(app)

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		bot.StopTelegramBot()
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Start Telegram bot
		if err := bot.StartTelegramBot(app); err != nil {
			log.Printf("Failed to start Telegram bot: %v", err)
		}

		// API routes
		se.Router.GET("/api/settings/{name}", api.GetSettingsHandler(app))
		se.Router.GET("/api/admin/summary", api.AdminSummaryHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/admin/events/{id}", api.AdminEventDetailsHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/admin/events/{id}/email", api.AdminEventEmailHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/admin/registrations/{id}/approve", api.AdminApproveRegistrationHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/admin/registrations/{id}/reject", api.AdminRejectRegistrationHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/admin/registrations/{id}/cancel", api.AdminCancelRegistrationHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/admin/groups/sync-memberships", api.AdminSyncGroupMembershipsHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/events/{slug}/register", api.RegisterEventHandler(app))
		se.Router.GET("/api/events/{slug}/status", api.EventStatusHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/events/{slug}/unsubscribe", api.EventUnsubscribeHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/events/accept", api.AcceptEventHandler(app))
		se.Router.POST("/api/telegram/generate-token", api.GenerateTelegramTokenHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/requests/submit", api.SubmitRequestHandler(app))
		se.Router.GET("/api/requests", api.ListRequestsHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/requests/{id}", api.GetRequestHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/requests/{id}/action", api.RequestActionHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/groups/{id}/members", api.GroupMembersHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/groups/{id}/guardians", api.GroupGuardiansHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/groups/{id}/requests-count", api.GroupRequestsCountHandler(app)).Bind(apis.RequireAuth())
		se.Router.GET("/api/groups/default-invite", api.DefaultGroupInviteHandler(app)).Bind(apis.RequireAuth())

		// Serve frontend
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
