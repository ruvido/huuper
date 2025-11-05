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

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Start Telegram bot
		if err := bot.StartTelegramBot(); err != nil {
			log.Printf("Failed to start Telegram bot: %v", err)
		}

		// API routes
		se.Router.POST("/api/telegram/link", api.LinkTelegramHandler(app)).Bind(apis.RequireAuth())

		// Serve frontend
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
