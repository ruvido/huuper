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
		se.Router.POST("/api/events/{slug}/register", api.RegisterEventHandler(app))
		se.Router.GET("/api/events/accept", api.AcceptEventHandler(app))
		se.Router.POST("/api/telegram/generate-token", api.GenerateTelegramTokenHandler(app)).Bind(apis.RequireAuth())

		// Serve frontend
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
