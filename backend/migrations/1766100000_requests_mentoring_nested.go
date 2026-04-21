package migrations

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Moves flat mentoring_notes / mentoring_done_at / mentoring_done_by fields
// stored on each request's data JSON into a nested mentoring object with a
// notes array (append-friendly):
//
//	data.mentoring = {
//	  notes: [{ text, at, by }],
//	  done_at, done_by
//	}
func init() {
	m.Register(func(app core.App) error {
		records, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			data := map[string]any{}
			if err := record.UnmarshalJSONField("data", &data); err != nil {
				return err
			}

			rawNotes, hasNotes := data["mentoring_notes"]
			rawDoneAt, hasDoneAt := data["mentoring_done_at"]
			rawDoneBy, hasDoneBy := data["mentoring_done_by"]
			if !hasNotes && !hasDoneAt && !hasDoneBy {
				continue
			}

			noteText := strings.TrimSpace(backendinternal.AnyToString(rawNotes))
			doneAt := strings.TrimSpace(backendinternal.AnyToString(rawDoneAt))
			doneBy := strings.TrimSpace(backendinternal.AnyToString(rawDoneBy))

			block := map[string]any{}
			if noteText != "" {
				entry := map[string]any{"text": noteText}
				if doneAt != "" {
					entry["at"] = doneAt
				}
				if doneBy != "" {
					entry["by"] = doneBy
				}
				block["notes"] = []any{entry}
			}
			if doneAt != "" {
				block["done_at"] = doneAt
			}
			if doneBy != "" {
				block["done_by"] = doneBy
			}

			if len(block) > 0 {
				data["mentoring"] = block
			}
			delete(data, "mentoring_notes")
			delete(data, "mentoring_done_at")
			delete(data, "mentoring_done_by")

			record.Set("data", data)
			if err := app.Save(record); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		return nil
	})
}

