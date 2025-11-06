package migrations

import (
	"encoding/json"
	"os"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Create settings collection
		settings := core.NewBaseCollection("settings")
		settings.ListRule = types.Pointer("") // Public read
		settings.ViewRule = types.Pointer("") // Public read
		settings.CreateRule = nil              // Admin only
		settings.UpdateRule = nil              // Admin only
		settings.DeleteRule = nil              // Admin only

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
				Max:      100,
			},
			&core.JSONField{
				Name:     "data",
				Required: true,
			},
		)

		// Add unique index on name field
		settings.AddIndex("idx_settings_name", true, "name", "")

		if err := app.Save(settings); err != nil {
			return err
		}

		// Seed telegram settings from .env
		telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		telegramBotName := os.Getenv("TELEGRAM_BOT_NAME")
		if telegramToken != "" && telegramBotName != "" {
			telegramData := map[string]string{
				"token": telegramToken,
				"name":  telegramBotName,
			}
			telegramJSON, _ := json.Marshal(telegramData)

			telegramRecord := core.NewRecord(settings)
			telegramRecord.Set("name", "telegram")
			telegramRecord.Set("data", string(telegramJSON))
			if err := app.Save(telegramRecord); err != nil {
				return err
			}
		}

		// Seed url settings from .env
		url := os.Getenv("URL")
		if url != "" {
			urlData := map[string]string{
				"address": url,
			}
			urlJSON, _ := json.Marshal(urlData)

			urlRecord := core.NewRecord(settings)
			urlRecord.Set("name", "url")
			urlRecord.Set("data", string(urlJSON))
			if err := app.Save(urlRecord); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete settings collection
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}
		return app.Delete(settings)
	})
}
