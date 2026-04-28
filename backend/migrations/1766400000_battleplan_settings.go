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

		existing, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'battleplan'",
			map[string]any{},
		)
		if err == nil && existing != nil {
			return nil
		}

		config := map[string]any{
			"priority": map[string]any{
				"label": "Nuovo Piano di Battaglia",
				"description": "Scegli un'esigenza profonda della tua vita. Sarà il faro del tuo cammino.\n- Cosa mi impedisce di vivere appieno?\n- Quale emozione o vuoto sento più forte in questo momento?\n- Come vorrei ricordare questo periodo tra un anno?",
			},
			"durations": []map[string]any{
				{"value": 30},
				{"value": 60, "default": true},
				{"value": 90},
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
			"visibility": []map[string]any{
				{"value": "group", "label": "Gruppo", "default": true},
				{"value": "public", "label": "Pubblico"},
			},
			"wizard": map[string]any{
				"intro": map[string]any{
					"title":  "Piano di Battaglia",
					"text":   "Una sfida personale a 30, 60 o 90 giorni. Una Priorità, quattro pilastri, routine quotidiane.",
					"button": "INIZIA",
				},
				"confirmation": map[string]any{
					"title":  "Pronto a partire",
					"text":   "Il tuo Piano di Battaglia è pronto.",
					"button": "Conferma",
				},
			},
		}

		record := core.NewRecord(settings)
		record.Set("name", "battleplan")
		record.Set("data", config)
		return app.Save(record)
	}, func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'battleplan'",
			map[string]any{},
		)
		if err == nil && record != nil {
			return app.Delete(record)
		}
		return nil
	})
}
