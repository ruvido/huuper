package migrations

import (
	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Backfill migration: convert the legacy single-copy priority shape
//
//	priority: { label, description }
//
// into the new variant shape that distinguishes the new-plan vs edit-plan copy:
//
//	priority: { new: { title, text }, edit: { title, text } }
//
// The conversion seeds both variants from the same source so existing copy is
// preserved verbatim; admins can split them later from the settings UI.
//
// Idempotent: skipped when priority.new already exists.
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

		data := backendinternal.ParseJSONMap(record.Get("data"))
		priority, _ := data["priority"].(map[string]any)
		if priority == nil {
			return nil
		}
		if _, hasNew := priority["new"].(map[string]any); hasNew {
			// Already migrated.
			return nil
		}
		label, _ := priority["label"].(string)
		description, _ := priority["description"].(string)
		if label == "" && description == "" {
			// Nothing meaningful to migrate; leave the record untouched so a
			// later seed migration can populate the canonical defaults.
			return nil
		}

		variant := map[string]any{"title": label, "text": description}
		data["priority"] = map[string]any{
			"new":  variant,
			"edit": variant,
		}
		record.Set("data", data)
		return app.Save(record)
	}, func(app core.App) error {
		// Best-effort reverse: collapse priority.new back to label/description
		// when present. Skip silently otherwise.
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'battleplan'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return err
		}
		data := backendinternal.ParseJSONMap(record.Get("data"))
		priority, _ := data["priority"].(map[string]any)
		if priority == nil {
			return nil
		}
		newVariant, ok := priority["new"].(map[string]any)
		if !ok {
			return nil
		}
		title, _ := newVariant["title"].(string)
		text, _ := newVariant["text"].(string)
		data["priority"] = map[string]any{
			"label":       title,
			"description": text,
		}
		record.Set("data", data)
		return app.Save(record)
	})
}
