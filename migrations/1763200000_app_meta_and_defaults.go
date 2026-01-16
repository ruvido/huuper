package migrations

import (
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	defaultAppTitle      = "App Title"
	defaultWelcomeMessage = "Welcome message"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}

		titleValue := strings.TrimSpace(os.Getenv("APP_NAME"))
		if titleValue == "" {
			titleValue = defaultAppTitle
		}

		titleRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'title'",
			map[string]any{},
		)
		if err == nil && titleRecord != nil {
			var titleData struct {
				Name string `json:"name"`
			}
			if err := titleRecord.UnmarshalJSONField("data", &titleData); err == nil && titleData.Name != "" {
				titleValue = titleData.Name
			}
		} else {
			titleRecord := core.NewRecord(settings)
			titleRecord.Set("name", "title")
			titleRecord.Set("data", map[string]string{"name": titleValue})
			if err := app.Save(titleRecord); err != nil {
				return err
			}
		}

		welcomeRecord, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'welcome'",
			map[string]any{},
		)
		if err != nil || welcomeRecord == nil {
			welcomeRecord := core.NewRecord(settings)
			welcomeRecord.Set("name", "welcome")
			welcomeRecord.Set("data", map[string]string{"content": defaultWelcomeMessage})
			if err := app.Save(welcomeRecord); err != nil {
				return err
			}
		}

		urlValue := strings.TrimSpace(os.Getenv("URL"))

		settingsModel := app.Settings()
		updated := false
		if settingsModel.Meta.AppName == "" || settingsModel.Meta.AppName == "Acme" {
			settingsModel.Meta.AppName = titleValue
			updated = true
		}
		if urlValue != "" && (settingsModel.Meta.AppURL == "" || settingsModel.Meta.AppURL == "http://localhost:8090") {
			settingsModel.Meta.AppURL = urlValue
			updated = true
		}

		if updated {
			if err := app.Save(settingsModel); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
