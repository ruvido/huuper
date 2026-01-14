package migrations

import (
	"encoding/json"
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
				&core.TextField{
					Name:     "data",
					Required: true,
					Max:      5000,
				},
			)
			if err := app.Save(settings); err != nil {
				return err
			}
		}

		// Seed bot messages from .env (required)
		msgWelcomeRaw := os.Getenv("MESSAGE_WELCOME")
		msgNotRegisteredRaw := os.Getenv("MESSAGE_NOT_REGISTERED")
		msgRegisteredRaw := os.Getenv("MESSAGE_REGISTERED")

		if msgWelcomeRaw == "" || msgNotRegisteredRaw == "" || msgRegisteredRaw == "" {
			return fmt.Errorf("MESSAGE_WELCOME, MESSAGE_NOT_REGISTERED, and MESSAGE_REGISTERED must be set in .env")
		}

		// Remove surrounding quotes and parse escape sequences
		msgWelcome := strings.Trim(msgWelcomeRaw, `"`)
		msgWelcome = strings.ReplaceAll(msgWelcome, `\n`, "\n")

		msgNotRegistered := strings.Trim(msgNotRegisteredRaw, `"`)
		msgNotRegistered = strings.ReplaceAll(msgNotRegistered, `\n`, "\n")

		msgRegistered := strings.Trim(msgRegisteredRaw, `"`)
		msgRegistered = strings.ReplaceAll(msgRegistered, `\n`, "\n")

		// Delete existing bot_messages record if it exists
		existingRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'bot_messages'",
			map[string]any{},
		)
		if err == nil && existingRecord != nil {
			if err := app.Delete(existingRecord); err != nil {
				return err
			}
		}

		// Create new bot_messages record with correct newlines
		botMessagesData := map[string]string{
			"welcome":        msgWelcome,
			"not_registered": msgNotRegistered,
			"registered":     msgRegistered,
		}
		botMessagesJSON, _ := json.Marshal(botMessagesData)

		botMessagesRecord := core.NewRecord(settings)
		botMessagesRecord.Set("name", "bot_messages")
		botMessagesRecord.Set("data", string(botMessagesJSON))
		if err := app.Save(botMessagesRecord); err != nil {
			return err
		}

		// Seed URL from .env (required)
		urlEnv := os.Getenv("URL")
		if urlEnv == "" {
			return fmt.Errorf("URL must be set in .env")
		}

		// Delete existing url record if it exists
		existingURLRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'url'",
			map[string]any{},
		)
		if err == nil && existingURLRecord != nil {
			if err := app.Delete(existingURLRecord); err != nil {
				return err
			}
		}

		// Create new url record
		urlData := map[string]string{
			"address": urlEnv,
		}
		urlJSON, _ := json.Marshal(urlData)

		urlRecord := core.NewRecord(settings)
		urlRecord.Set("name", "url")
		urlRecord.Set("data", string(urlJSON))
		if err := app.Save(urlRecord); err != nil {
			return err
		}

		// Seed Telegram settings from .env (required)
		telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		telegramName := os.Getenv("TELEGRAM_BOT_NAME")

		if telegramToken == "" || telegramName == "" {
			return fmt.Errorf("TELEGRAM_BOT_TOKEN and TELEGRAM_BOT_NAME must be set in .env")
		}

		// Check if telegram record already exists
		existingTelegramRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'telegram'",
			map[string]any{},
		)

		// Only create if it doesn't exist
		if err != nil || existingTelegramRecord == nil {
			telegramData := map[string]string{
				"token": telegramToken,
				"name":  telegramName,
			}
			telegramJSON, _ := json.Marshal(telegramData)

			telegramRecord := core.NewRecord(settings)
			telegramRecord.Set("name", "telegram")
			telegramRecord.Set("data", string(telegramJSON))
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
		users.Fields.RemoveById("status")
		return app.Save(users)
	})
}
