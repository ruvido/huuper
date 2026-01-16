package migrations

import (
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Add status field to users collection
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"active", "pending"},
		})

		if err := app.Save(users); err != nil {
			return err
		}

		// Set default status to "active" for all existing users
		records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			record.Set("status", "active")
			if err := app.Save(record); err != nil {
				return err
			}
		}

		// Create settings collection if it doesn't exist
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			// Collection doesn't exist, create it
			settings = core.NewBaseCollection("settings")
			settings.Fields.Add(
				&core.AutodateField{
					Name:     "created",
					OnCreate: true,
				},
				&core.AutodateField{
					Name:     "updated",
					OnCreate: true,
					OnUpdate: true,
				},
				&core.TextField{
					Name:     "name",
					Required: true,
					Max:      255,
				},
				&core.JSONField{
					Name:     "data",
					Required: true,
				},
			)
			if err := app.Save(settings); err != nil {
				return err
			}
		}

		// Seed bot messages with defaults if missing
		existingRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'bot_messages'",
			map[string]any{},
		)
		if err != nil || existingRecord == nil {
			// Create new bot_messages record with correct newlines
			botMessagesData := map[string]string{
				"welcome": "Welcome! This group requires registration.\n\nPlease sign up here:\n{url}",
				"warning": "This bot doesn't reply to messages.\n\nPlease sign up on the web app:\n{url}",
			}

			botMessagesRecord := core.NewRecord(settings)
			botMessagesRecord.Set("name", "bot_messages")
			botMessagesRecord.Set("data", botMessagesData)
			if err := app.Save(botMessagesRecord); err != nil {
				return err
			}
		}

		// Seed URL from .env (required)
		existingURLRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'url'",
			map[string]any{},
		)
		if err != nil || existingURLRecord == nil {
			urlEnv := os.Getenv("URL")
			if urlEnv == "" {
				return fmt.Errorf("missing required env vars: URL\n\n.env.example:\n%s", envExample)
			}

			urlData := map[string]string{
				"address": urlEnv,
			}

			urlRecord := core.NewRecord(settings)
			urlRecord.Set("name", "url")
			urlRecord.Set("data", urlData)
			if err := app.Save(urlRecord); err != nil {
				return err
			}
		}

		// Seed Telegram settings from .env (required)
		existingTelegramRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'telegram'",
			map[string]any{},
		)
		if err != nil || existingTelegramRecord == nil {
			telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
			telegramName := os.Getenv("TELEGRAM_BOT_NAME")

			if telegramToken == "" || telegramName == "" {
				missing := []string{}
				if telegramToken == "" {
					missing = append(missing, "TELEGRAM_BOT_TOKEN")
				}
				if telegramName == "" {
					missing = append(missing, "TELEGRAM_BOT_NAME")
				}
				return fmt.Errorf("missing required env vars: %s\n\n.env.example:\n%s", strings.Join(missing, ", "), envExample)
			}

			telegramData := map[string]string{
				"token": telegramToken,
				"name":  telegramName,
			}

			telegramRecord := core.NewRecord(settings)
			telegramRecord.Set("name", "telegram")
			telegramRecord.Set("data", telegramData)
			if err := app.Save(telegramRecord); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete url record from settings
		urlRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'url'",
			map[string]any{},
		)
		if err == nil && urlRecord != nil {
			app.Delete(urlRecord)
		}

		// Downgrade: delete bot_messages record from settings
		botMessagesRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'bot_messages'",
			map[string]any{},
		)
		if err == nil && botMessagesRecord != nil {
			app.Delete(botMessagesRecord)
		}

		// Remove status field from users collection
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		statusField := users.Fields.GetByName("status")
		if statusField != nil {
			users.Fields.RemoveById(statusField.GetId())
		}
		return app.Save(users)
	})
}

const envExample = `ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme123456
PORT=8090
APP_NAME=App Title

TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_BOT_NAME=@your_bot_name

URL=http://localhost:8090
`
