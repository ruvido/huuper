package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Convert events.cadence from a free-text field to a SelectField with the
// fixed enum the cadence parser already understands. Keeps PB admin UX
// honest: no typos, dropdown with the actual valid values.
func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		if field := events.Fields.GetByName("cadence"); field != nil {
			events.Fields.RemoveById(field.GetId())
		}
		events.Fields.Add(&core.SelectField{
			Name:      "cadence",
			Required:  false,
			MaxSelect: 1,
			Values:    cadenceValues(),
		})

		if err := app.Save(events); err != nil {
			return err
		}

		// Backfill any record whose cadence got cleared during the field swap.
		records, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.GetString("cadence") == "" {
				record.Set("cadence", "once")
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
		if field := events.Fields.GetByName("cadence"); field != nil {
			events.Fields.RemoveById(field.GetId())
		}
		events.Fields.Add(&core.TextField{
			Name:     "cadence",
			Required: false,
			Max:      50,
		})
		return app.Save(events)
	})
}

func cadenceValues() []string {
	weekdays := []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	out := []string{"once"}
	for _, d := range weekdays {
		out = append(out, "weekly:"+d)
	}
	for _, n := range []string{"1st", "2nd", "3rd", "4th", "5th", "last"} {
		for _, d := range weekdays {
			out = append(out, "monthly:"+n+"-"+d)
		}
	}
	return out
}
