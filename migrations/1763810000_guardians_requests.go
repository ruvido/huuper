package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		guardians, err := app.FindCollectionByNameOrId("guardians")
		if err != nil {
			return err
		}

		guardians.Fields.Add(
			&core.RelationField{
				Name:         "request",
				Required:     false,
				CollectionId: requests.Id,
				MaxSelect:    1,
			},
		)

		userField := guardians.Fields.GetByName("user")
		if userField != nil {
			guardians.Fields.RemoveById(userField.GetId())
		}

		guardians.ListRule = types.Pointer("@request.auth.id != '' && (guardian = @request.auth.id || group.leader = @request.auth.id)")
		guardians.ViewRule = types.Pointer("@request.auth.id != '' && (guardian = @request.auth.id || group.leader = @request.auth.id)")
		guardians.CreateRule = types.Pointer("@request.auth.id != '' && group.leader = @request.auth.id")
		guardians.UpdateRule = types.Pointer("@request.auth.id != '' && (guardian = @request.auth.id || group.leader = @request.auth.id)")

		guardians.AddIndex("idx_guardians_request", true, "request", "")
		guardians.RemoveIndex("idx_guardians_user")

		if err := app.Save(guardians); err != nil {
			return err
		}

		requests.ListRule = types.Pointer("@request.auth.id != '' && (group.leader = @request.auth.id || (id ?= @collection.guardians.request && @collection.guardians.guardian = @request.auth.id))")
		requests.ViewRule = types.Pointer("@request.auth.id != '' && (group.leader = @request.auth.id || (id ?= @collection.guardians.request && @collection.guardians.guardian = @request.auth.id))")

		return app.Save(requests)
	}, func(app core.App) error {
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		guardians, err := app.FindCollectionByNameOrId("guardians")
		if err != nil {
			return err
		}

		guardians.Fields.Add(
			&core.RelationField{
				Name:         "user",
				Required:     false,
				CollectionId: users.Id,
				MaxSelect:    1,
			},
		)

		requestField := guardians.Fields.GetByName("request")
		if requestField != nil {
			guardians.Fields.RemoveById(requestField.GetId())
		}

		guardians.ListRule = nil
		guardians.ViewRule = nil
		guardians.CreateRule = nil
		guardians.UpdateRule = nil

		guardians.AddIndex("idx_guardians_user", true, "user", "")
		guardians.RemoveIndex("idx_guardians_request")

		if err := app.Save(guardians); err != nil {
			return err
		}

		requests, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return err
		}

		requests.ListRule = types.Pointer("@request.auth.id != '' && group.leader = @request.auth.id")
		requests.ViewRule = types.Pointer("@request.auth.id != '' && group.leader = @request.auth.id")

		return app.Save(requests)
	})
}
