package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'battleplan'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return err
		}

		record.Set("data", map[string]any{
			"priority": map[string]any{
				"label": "Nuovo Piano di Battaglia",
				"description": "Scegli un'esigenza profonda della tua vita. Sarà il faro del tuo cammino.\n- Cosa mi impedisce di vivere appieno?\n- Quale emozione o vuoto sento più forte in questo momento?\n- Come vorrei ricordare questo periodo tra un anno?",
			},
			"pillars": []map[string]any{
				{"key": "interiorita", "label": "Interiorità", "description": ""},
				{"key": "relazioni", "label": "Relazioni", "description": ""},
				{"key": "risorse", "label": "Risorse", "description": ""},
				{"key": "salute", "label": "Salute", "description": ""},
			},
			"cadences": []map[string]any{
				{"type": "daily", "label": "Ogni giorno", "default": true},
				{"type": "specific_days", "label": "Giorni specifici"},
				{"type": "times_per_week", "label": "Volte a settimana"},
			},
			"durations": []map[string]any{
				{"value": 30},
				{"value": 60, "default": true},
				{"value": 90},
			},
			"visibility": []map[string]any{
				{"value": "group", "label": "Gruppo", "default": true},
				{"value": "public", "label": "Pubblico"},
			},
			"wizard": map[string]any{
				"confirmation": map[string]any{
					"button": "Conferma",
					"text":   "Il tuo Piano di Battaglia è pronto.",
					"title":  "Pronto a partire",
				},
				"intro": map[string]any{
					"show":   false,
					"button": "INIZIA",
					"text":   "Una sfida personale a 30, 60 o 90 giorni. Una Priorità, quattro pilastri, routine quotidiane.",
					"title":  "Piano di Battaglia",
				},
			},
		})

		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}
