package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}
		record, _ := app.FindFirstRecordByFilter(
			"settings",
			"name = 'eventflow'",
			map[string]any{},
		)
		if record == nil {
			record = core.NewRecord(settings)
			record.Set("name", "eventflow")
		}
		record.Set("data", defaultEventflowV2())
		if err := app.Save(record); err != nil {
			return err
		}

		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		records, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return err
		}
		previousTypes := make(map[string]string, len(records))
		for _, record := range records {
			previousTypes[record.Id] = record.GetString("type")
		}
		events.RemoveIndex("idx_events_type_date")
		if field := events.Fields.GetByName("type"); field != nil {
			events.Fields.RemoveById(field.GetId())
		}
		events.Fields.Add(&core.TextField{
			Name:     "type",
			Required: true,
			Max:      80,
		})
		events.AddIndex("idx_events_type_date", false, "type, event_date", "")
		if err := app.Save(events); err != nil {
			return err
		}
		for _, record := range records {
			value := previousTypes[record.Id]
			if value == "" {
				value = "meetup"
			}
			record.Set("type", value)
			if err := app.Save(record); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		return nil
	})
}

func defaultEventflowV2() map[string]any {
	return map[string]any{
		"types": []map[string]any{
			{
				"value":       "call",
				"label":       "Call",
				"description": "Online call.",
				"creators":    []string{"admin", "assistant"},
				"required": map[string]bool{
					"title":    false,
					"url":      true,
					"location": false,
					"group":    false,
					"end_date": false,
				},
				"registration": map[string]bool{
					"enabled":  true,
					"approval": false,
				},
			},
			{
				"value":       "meetup",
				"label":       "Meetup",
				"description": "In-person meetup.",
				"creators":    []string{"admin", "assistant"},
				"required": map[string]bool{
					"title":    false,
					"url":      false,
					"location": true,
					"group":    false,
					"end_date": false,
				},
				"registration": map[string]bool{
					"enabled":  true,
					"approval": false,
				},
			},
		},
	}
}
