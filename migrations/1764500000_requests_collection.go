package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		if existing, err := app.FindCollectionByNameOrId("requests"); err == nil && existing != nil {
			return nil
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		requests := core.NewBaseCollection("requests")
		requests.ListRule = nil
		requests.ViewRule = nil
		requests.CreateRule = nil
		requests.UpdateRule = nil
		requests.DeleteRule = nil

		requests.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.RelationField{
				Name:         "guardian",
				Required:     false,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
			&core.JSONField{
				Name:     "data",
				Required: true,
			},
		)

		requests.AddIndex(
			"idx_requests_guardian",
			false,
			"guardian",
			"guardian != ''",
		)

		return app.Save(requests)
	}, func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil || requests == nil {
			return nil
		}

		return app.Delete(requests)
	})
}
