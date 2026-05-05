package migrations

import (
	"strconv"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Convert events.count from NumberField to SelectField with the four values
// the wizard actually offers (1 / 3 / 6 / 12). The Go layer reads this back
// as a string and parses to int via strconv.Atoi at the call site.
func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		// Snapshot current numeric counts so we can re-seed after the type swap.
		records, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return err
		}
		previous := make(map[string]int, len(records))
		for _, r := range records {
			previous[r.Id] = r.GetInt("count")
		}

		// Drop the NumberField column first and persist, otherwise SQLite keeps
		// the NUMERIC column affinity and silently coerces the new string values
		// back to floats — which then fail SelectField validation as "1.0".
		if field := events.Fields.GetByName("count"); field != nil {
			events.Fields.RemoveById(field.GetId())
			if err := app.Save(events); err != nil {
				return err
			}
		}
		events.Fields.Add(&core.SelectField{
			Name:      "count",
			Required:  false,
			MaxSelect: 1,
			Values:    []string{"1", "3", "6", "12"},
		})
		if err := app.Save(events); err != nil {
			return err
		}

		// Backfill: snap each existing count to the nearest allowed value.
		// Re-fetch records after the schema swap so the field cache reflects the
		// new SelectField (otherwise stored values keep their old numeric form).
		fresh, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, r := range fresh {
			r.Set("count", strconv.Itoa(snapCount(previous[r.Id])))
			if err := app.Save(r); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		if field := events.Fields.GetByName("count"); field != nil {
			events.Fields.RemoveById(field.GetId())
		}
		events.Fields.Add(&core.NumberField{
			Name:     "count",
			Required: false,
			Min:      pointerToFloatCount(1),
			Max:      pointerToFloatCount(52),
		})
		return app.Save(events)
	})
}

func snapCount(value int) int {
	switch {
	case value <= 1:
		return 1
	case value <= 3:
		return 3
	case value <= 6:
		return 6
	default:
		return 12
	}
}

func pointerToFloatCount(v float64) *float64 { return &v }
