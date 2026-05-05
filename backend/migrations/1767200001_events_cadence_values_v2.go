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
		field := events.Fields.GetByName("cadence")
		if field == nil {
			return nil
		}
		selectField, ok := field.(*core.SelectField)
		if !ok {
			return nil
		}
		selectField.Values = eventCadenceValuesV2()
		return app.Save(events)
	}, func(app core.App) error {
		return nil
	})
}

func eventCadenceValuesV2() []string {
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
