package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		regions, err := app.FindCollectionByNameOrId("regions")
		if err != nil {
			return err
		}

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		requests := core.NewBaseCollection("requests")
		requests.ListRule = types.Pointer("@request.auth.id != '' && group.leader = @request.auth.id")
		requests.ViewRule = types.Pointer("@request.auth.id != '' && group.leader = @request.auth.id")
		requests.CreateRule = types.Pointer("")
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
			&core.TextField{
				Name:     "name",
				Required: true,
				Max:      200,
			},
			&core.EmailField{
				Name:     "email",
				Required: true,
			},
			&core.TextField{
				Name:     "motivation",
				Required: true,
				Max:      2000,
			},
			&core.TextField{
				Name:     "birth_year",
				Required: true,
				Max:      4,
			},
			&core.RelationField{
				Name:         "region",
				Required:     true,
				CollectionId: regions.Id,
				MaxSelect:    1,
			},
			&core.SelectField{
				Name:     "civil_status",
				Required: true,
				Values:   []string{"single", "married", "other"},
			},
			&core.SelectField{
				Name:     "status",
				Required: true,
				Values:   []string{"0-pending", "1-accepted", "2-assigned", "3-approved", "9-rejected"},
			},
			&core.RelationField{
				Name:         "group",
				Required:     false,
				CollectionId: groups.Id,
				MaxSelect:    1,
			},
		)

		requests.AddIndex("idx_requests_email", true, "email", "")

		if err := app.Save(requests); err != nil {
			return err
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		users.Fields.Add(
			&core.DateField{
				Name:     "request_at",
				Required: false,
			},
		)

		return app.Save(users)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		requestAt := users.Fields.GetByName("request_at")
		if requestAt != nil {
			users.Fields.RemoveById(requestAt.GetId())
			if err := app.Save(users); err != nil {
				return err
			}
		}

		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		return app.Delete(requests)
	})
}
