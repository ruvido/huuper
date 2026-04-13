package main

import (
	"log"
	"sync"
	"time"

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
	var watchOnce sync.Once
	var skeletonDir string
	var publicDir string
	var frontendDev bool
	var disableTelegramBot bool

	app.RootCmd.PersistentFlags().StringVar(
		&skeletonDir,
		"skeletonDir",
		defaultSkeletonDir,
		"the directory with frontend source files",
	)
	app.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir,
		"the directory to serve generated frontend files",
	)
	app.RootCmd.PersistentFlags().BoolVar(
		&frontendDev,
		"frontend-dev",
		false,
		"build frontend from skeleton and enable live reload on changes",
	)
	app.RootCmd.PersistentFlags().BoolVar(
		&disableTelegramBot,
		"disable-telegram-bot",
		false,
		"skip starting the Telegram bot",
	)
	hooks.RegisterSettingsValidation(app)
	hooks.RegisterGroupsValidation(app)
	hooks.RegisterUsersNormalization(app)

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		bot.StopTelegramBot()
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if frontendDev {
			liveReload := newLiveReloadHub()

			if err := buildFrontend(skeletonDir, publicDir, true); err != nil {
				return err
			}

			se.Router.GET("/_dev/live-reload", func(e *core.RequestEvent) error {
				e.Response.Header().Set("Content-Type", "text/event-stream")
				e.Response.Header().Set("Cache-Control", "no-cache")
				e.Response.Header().Set("Connection", "keep-alive")

				flusher, ok := e.Response.(interface{ Flush() })
				if !ok {
					return e.InternalServerError("streaming unsupported", nil)
				}

				ch := liveReload.Subscribe()
				defer liveReload.Unsubscribe(ch)

				_, _ = e.Response.Write([]byte(": connected\n\n"))
				flusher.Flush()

				ctx := e.Request.Context()
				ticker := time.NewTicker(30 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return nil
					case <-ticker.C:
						_, _ = e.Response.Write([]byte(": ping\n\n"))
						flusher.Flush()
					case <-ch:
						_, _ = e.Response.Write([]byte("event: reload\ndata: frontend-updated\n\n"))
						flusher.Flush()
					}
				}
			})

			var watchErr error
			watchOnce.Do(func() {
				watchErr = startFrontendWatcher(skeletonDir, publicDir, true, func() {
					liveReload.Publish()
				})
				if watchErr == nil {
					log.Printf("frontend watcher active: %s -> %s", skeletonDir, publicDir)
				}
			})
			if watchErr != nil {
				return watchErr
			}
		}

		// Start Telegram bot unless explicitly disabled.
		if !disableTelegramBot {
			if err := bot.StartTelegramBot(app); err != nil {
				log.Printf("Failed to start Telegram bot: %v", err)
			}
		}

		api.RegisterRoutes(app, se)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
