package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"members/backend/api"
	"members/backend/bot"
	"members/backend/internal/hooks"
	_ "members/backend/migrations"
)

func init() {
	// Load .env file BEFORE migrations run
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	app := pocketbase.New()
	hooks.RegisterSettingsValidation(app)
	hooks.RegisterGroupsValidation(app)
	hooks.RegisterUsersNormalization(app)

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		bot.StopTelegramBot()
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Start Telegram bot
		if err := bot.StartTelegramBot(app); err != nil {
			log.Printf("Failed to start Telegram bot: %v", err)
		}

		api.RegisterRoutes(app, se)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
