package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Refactor events schema:
//   - drop `series` (recurrence is now described by cadence + count on a single record)
//   - add `cadence` (once | weekly:X | monthly:Nth-X | monthly:last-X)
//   - add `count` (int, total occurrences when cadence != once)
//   - add `end_date` (optional datetime, lets a single occurrence span multiple days)
//   - add `location` (optional string, mainly used by meetup)
//   - drop "rally" from the type select; existing rally records become meetup
//     (national/open semantics now expressed as meetup with no group)
func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		dirty := false

		// idx_events_series was created by 1766700000 alongside the series
		// column. Drop the index BEFORE removing its column or PB tries to
		// re-create it on save and fails with "no such column: series".
		events.RemoveIndex("idx_events_series")

		if field := events.Fields.GetByName("series"); field != nil {
			events.Fields.RemoveById(field.GetId())
			dirty = true
		}

		if events.Fields.GetByName("cadence") == nil {
			events.Fields.Add(&core.TextField{
				Name:     "cadence",
				Required: false,
				Max:      50,
			})
			dirty = true
		}

		if events.Fields.GetByName("count") == nil {
			events.Fields.Add(&core.NumberField{
				Name:     "count",
				Required: false,
				Min:      pointerToFloat(1),
				Max:      pointerToFloat(52),
			})
			dirty = true
		}

		if events.Fields.GetByName("end_date") == nil {
			events.Fields.Add(&core.DateField{
				Name:     "end_date",
				Required: false,
			})
			dirty = true
		}

		if events.Fields.GetByName("location") == nil {
			events.Fields.Add(&core.TextField{
				Name:     "location",
				Required: false,
				Max:      300,
			})
			dirty = true
		}

		if typeField := events.Fields.GetByName("type"); typeField != nil {
			if sel, ok := typeField.(*core.SelectField); ok {
				sel.Values = []string{"call", "meetup"}
				dirty = true
			}
		}

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
			recordDirty := false
			if record.GetString("type") == "rally" {
				record.Set("type", "meetup")
				recordDirty = true
			}
			if record.GetString("cadence") == "" {
				record.Set("cadence", "once")
				recordDirty = true
			}
			if record.GetInt("count") == 0 {
				record.Set("count", 1)
				recordDirty = true
			}
			if recordDirty {
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
		for _, name := range []string{"cadence", "count", "end_date", "location"} {
			if field := events.Fields.GetByName(name); field != nil {
				events.Fields.RemoveById(field.GetId())
			}
		}
		if events.Fields.GetByName("series") == nil {
			events.Fields.Add(&core.TextField{
				Name:     "series",
				Required: false,
				Max:      50,
			})
			events.AddIndex("idx_events_series", false, "series", "series != ''")
		}
		if typeField := events.Fields.GetByName("type"); typeField != nil {
			if sel, ok := typeField.(*core.SelectField); ok {
				sel.Values = []string{"rally", "call", "meetup"}
			}
		}
		return app.Save(events)
	})
}

func pointerToFloat(v float64) *float64 { return &v }
