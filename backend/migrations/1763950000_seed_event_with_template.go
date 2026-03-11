package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		eventSlug := "evento-test"
		existingEvent, _ := app.FindFirstRecordByFilter(
			"events",
			"slug = {:slug}",
			map[string]any{"slug": eventSlug},
		)
		if existingEvent != nil {
			return nil
		}

		eventRecord := core.NewRecord(events)
		eventRecord.Set("title", "Evento Test")
		eventRecord.Set("slug", eventSlug)
		eventRecord.Set("active", true)
		eventRecord.Set("event_date", types.NowDateTime().AddDate(0, 0, 7))
		eventRecord.Set("data", map[string]any{
			"messages": map[string]any{
				"invalid_event":      "Evento non valido.",
				"invalid_email":      "Email non valida.",
				"error_generic":      "Registrazione non riuscita.",
				"event_closed":       "Mi dispiace, le iscrizioni sono chiuse per questo evento.",
				"already_submitted":  "Registrazione fallita, questa email è già stata usata.",
				"success_with_email": "Controlla la tua email per la conferma.",
				"success_no_email":   "Grazie, abbiamo ricevuto la tua iscrizione.",
			},
		})

		return app.Save(eventRecord)
	}, func(app core.App) error {
		existingEvent, _ := app.FindFirstRecordByFilter(
			"events",
			"slug = {:slug}",
			map[string]any{"slug": "evento-test"},
		)
		if existingEvent != nil {
			if err := app.Delete(existingEvent); err != nil {
				return err
			}
		}

		return nil
	})
}
