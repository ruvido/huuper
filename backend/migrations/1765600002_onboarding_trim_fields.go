package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		record, err := app.FindFirstRecordByFilter(
			"settings",
			"name = 'onboarding'",
			map[string]any{},
		)
		if err != nil || record == nil {
			return nil
		}

		data := map[string]any{}
		_ = record.UnmarshalJSONField("data", &data)
		updated := normalizeOnboardingSettings(data)
		if len(updated) == 0 {
			return nil
		}
		record.Set("data", updated)
		return app.Save(record)
	}, func(app core.App) error {
		return nil
	})
}

func normalizeOnboardingSettings(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	startPage := normalizeOnboardingPage(data["start_page"])
	confirmation := normalizeOnboardingPage(data["confirmation"])
	steps := normalizeOnboardingSteps(data["steps"])
	expected := defaultOnboardingSteps()

	if len(steps) == len(expected) {
		match := true
		for i := range steps {
			if strings.TrimSpace(toString(steps[i]["field"])) != strings.TrimSpace(toString(expected[i]["field"])) {
				match = false
				break
			}
		}
		if match {
			return nil
		}
	}

	out := map[string]any{
		"steps": expected,
	}
	if len(startPage) > 0 {
		out["start_page"] = startPage
	}
	if len(confirmation) > 0 {
		out["confirmation"] = confirmation
	}
	return out
}

func normalizeOnboardingPage(raw any) map[string]any {
	page, ok := raw.(map[string]any)
	if !ok || page == nil {
		return nil
	}

	out := map[string]any{}
	if value := strings.TrimSpace(toString(page["title"])); value != "" {
		out["title"] = value
	}
	if value := strings.TrimSpace(toString(page["text"])); value != "" {
		out["text"] = value
	}
	if value := strings.TrimSpace(toString(page["button"])); value != "" {
		out["button"] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeOnboardingSteps(raw any) []map[string]any {
	steps, ok := raw.([]any)
	if !ok || len(steps) == 0 {
		return nil
	}

	out := make([]map[string]any, 0, len(steps))
	for _, item := range steps {
		step, ok := item.(map[string]any)
		if !ok {
			continue
		}
		field := strings.TrimSpace(toString(step["field"]))
		if field == "" {
			continue
		}
		out = append(out, map[string]any{
			"field": field,
			"title": strings.TrimSpace(toString(step["title"])),
			"label": strings.TrimSpace(toString(step["label"])),
		})
	}
	return out
}

func defaultOnboardingSteps() []map[string]any {
	return []map[string]any{
		{"field": "work", "title": "In che campo lavori?"},
		{"field": "skills", "title": "Le tue skill", "label": "Nel lavoro o tempo libero, cosa sai fare con le mani?"},
		{"field": "interests", "title": "I tuoi interessi", "label": "Cosa ti appassiona? Quali sono i tuoi hobby?"},
		{"field": "sports", "title": "I tuoi sport", "label": "Dove ti piace metterti alla prova?"},
		{"field": "avatar", "title": "La tua foto!", "label": "Fatti vedere, così possiamo riconoscerti!"},
	}
}
