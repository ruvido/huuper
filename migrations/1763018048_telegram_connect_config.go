package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get settings collection
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		// Create telegram_connect config if missing
		existing, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'telegram_connect'",
			map[string]any{},
		)
		if err == nil && existing != nil {
			return nil
		}

		config := map[string]any{
			"title":       "Connetti Telegram",
			"main_text":   "Per accedere ai gruppi devi connettere il tuo account Telegram.",
			"description": "Clicca sul pulsante qui sotto per collegare il tuo profilo Telegram e accedere ai contenuti riservati della community.",
			"button":      "CONNETTI TELEGRAM",
			"loading":     "CONNESSIONE...",
		}

		record := core.NewRecord(settings)
		record.Set("name", "telegram_connect")
		record.Set("data", config)
		if err := app.Save(record); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete telegram_connect config
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'telegram_connect'",
			map[string]any{},
		)
		if err == nil && record != nil {
			app.Delete(record)
		}
		return nil
	})
}
