package migrations

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		if requests.Fields.GetByName("email") == nil {
			requests.Fields.Add(&core.TextField{
				Name:     "email",
				Required: false,
				Max:      255,
			})
		}

		if requests.Fields.GetByName("group") == nil {
			requests.Fields.Add(&core.RelationField{
				Name:         "group",
				Required:     false,
				CollectionId: groups.Id,
				MaxSelect:    1,
			})
		}

		if err := app.Save(requests); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		seenEmails := make(map[string]struct{}, len(records))
		for _, record := range records {
			data := map[string]any{}
			_ = record.UnmarshalJSONField("data", &data)

			email := strings.ToLower(strings.TrimSpace(record.GetString("email")))
			if email == "" {
				if raw, ok := data["email"].(string); ok {
					email = strings.ToLower(strings.TrimSpace(raw))
				}
			}

			delete(data, "email")
			record.Set("data", data)

			if email != "" {
				if _, exists := seenEmails[email]; exists {
					return fmt.Errorf("duplicate requests.email detected: %s", email)
				}
				seenEmails[email] = struct{}{}
				record.Set("email", email)
			}

			if err := app.Save(record); err != nil {
				return err
			}
		}

		requests.AddIndex(
			"idx_requests_email",
			true,
			"email",
			"email != ''",
		)
		requests.AddIndex(
			"idx_requests_group",
			false,
			"`group`",
			"`group` != ''",
		)

		return app.Save(requests)
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return nil
		}

		requests.RemoveIndex("idx_requests_email")
		requests.RemoveIndex("idx_requests_group")

		if field := requests.Fields.GetByName("email"); field != nil {
			requests.Fields.RemoveById(field.GetId())
		}
		if field := requests.Fields.GetByName("group"); field != nil {
			requests.Fields.RemoveById(field.GetId())
		}

		return app.Save(requests)
	})
}
