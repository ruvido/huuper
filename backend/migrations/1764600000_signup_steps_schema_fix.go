package migrations

import (
	"strings"

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
			"name = 'signup'",
			map[string]any{},
		)
		if record == nil {
			record = core.NewRecord(settings)
			record.Set("name", "signup")
		}

		data := map[string]any{}
		_ = record.UnmarshalJSONField("data", &data)
		if nested, ok := data["data"].(map[string]any); ok && nested != nil {
			data = nested
		}

		if strings.TrimSpace(toString(data["title"])) == "" {
			data["title"] = "Candidatura Realmen"
		}
		if strings.TrimSpace(toString(data["submit_label"])) == "" {
			data["submit_label"] = "Invia richiesta"
		}

		steps := normalizeSignupSteps(data)
		if len(steps) == 0 {
			steps = defaultSignupSteps()
		}

		delete(data, "fields")
		delete(data, "request_defaults")
		data["steps"] = steps
		record.Set("data", data)

		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}

func normalizeSignupSteps(data map[string]any) []map[string]any {
	if data == nil {
		return nil
	}

	if rawSteps, ok := data["steps"].([]any); ok && len(rawSteps) > 0 {
		out := make([]map[string]any, 0, len(rawSteps))
		for _, raw := range rawSteps {
			step, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			field := normalizeSignupFieldKey(step["field"])
			if field == "" {
				continue
			}
			entry := map[string]any{"field": field}
			if label := strings.TrimSpace(toString(step["label"])); label != "" {
				entry["label"] = label
			}
			out = append(out, entry)
		}
		return out
	}

	if rawFields, ok := data["fields"].([]any); ok && len(rawFields) > 0 {
		out := make([]map[string]any, 0, len(rawFields))
		for _, raw := range rawFields {
			fieldCfg, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			field := normalizeSignupFieldKey(fieldCfg["key"])
			if field == "" {
				continue
			}
			entry := map[string]any{"field": field}
			if label := strings.TrimSpace(toString(fieldCfg["label"])); label != "" {
				entry["label"] = label
			}
			out = append(out, entry)
		}
		return out
	}

	return nil
}

func normalizeSignupFieldKey(value any) string {
	field := strings.TrimSpace(toString(value))
	if field == "name" {
		return "full_name"
	}
	return field
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func defaultSignupSteps() []map[string]any {
	return []map[string]any{
		{"field": "full_name", "label": "Nome e cognome"},
		{"field": "email", "label": "Email"},
		{"field": "mobile", "label": "Cellulare"},
		{"field": "region", "label": "Regione"},
		{"field": "birth_year", "label": "Anno di nascita"},
		{"field": "marital_status", "label": "Stato relazionale"},
		{"field": "children", "label": "Figli"},
		{"field": "motivation", "label": "Motivazione"},
	}
}
