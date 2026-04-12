package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		records, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			value := strings.TrimSpace(record.GetString("type"))
			switch value {
			case "general", "default":
				record.Set("type", "local")
			case "local", "public", "private", "":
				continue
			default:
				record.Set("type", "private")
			}
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		return nil
	})
}
