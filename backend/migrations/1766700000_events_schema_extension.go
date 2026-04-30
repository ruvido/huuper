package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		dirty := false

		if events.Fields.GetByName("type") == nil {
			events.Fields.Add(&core.SelectField{
				Name:      "type",
				Required:  true,
				MaxSelect: 1,
				Values:    []string{"rally", "call", "meetup"},
			})
			dirty = true
		}

		if events.Fields.GetByName("group") == nil {
			events.Fields.Add(&core.RelationField{
				Name:         "group",
				Required:     false,
				CollectionId: groups.Id,
				MaxSelect:    1,
			})
			dirty = true
		}

		if events.Fields.GetByName("created_by") == nil {
			events.Fields.Add(&core.RelationField{
				Name:         "created_by",
				Required:     false,
				CollectionId: users.Id,
				MaxSelect:    1,
			})
			dirty = true
		}

		if events.Fields.GetByName("url") == nil {
			events.Fields.Add(&core.TextField{
				Name:     "url",
				Required: false,
				Max:      500,
			})
			dirty = true
		}

		if events.Fields.GetByName("series") == nil {
			events.Fields.Add(&core.TextField{
				Name:     "series",
				Required: false,
				Max:      50,
			})
			dirty = true
		}

		events.AddIndex("idx_events_type_date", false, "type, event_date", "")
		events.AddIndex("idx_events_group_date", false, "group, event_date", "")
		events.AddIndex("idx_events_series", false, "series", "series != ''")

		if dirty {
			if err := app.Save(events); err != nil {
				return err
			}
		}

		records, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.GetString("type") == "" {
				record.Set("type", "rally")
				if err := app.Save(record); err != nil {
					return err
				}
			}
		}

		return nil
	}, func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		for _, name := range []string{"type", "group", "created_by", "url", "series"} {
			if field := events.Fields.GetByName(name); field != nil {
				events.Fields.RemoveById(field.GetId())
			}
		}
		return app.Save(events)
	})
}
