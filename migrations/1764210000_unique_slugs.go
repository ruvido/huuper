package migrations

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		if err := ensureUniqueSlug(app, "events"); err != nil {
			return err
		}
		if err := ensureUniqueSlug(app, "templates"); err != nil {
			return err
		}

		events.AddIndex("idx_events_slug", true, "slug", "slug != ''")
		if err := app.Save(events); err != nil {
			return err
		}

		templates.AddIndex("idx_templates_slug", true, "slug", "slug != ''")
		return app.Save(templates)
	}, func(app core.App) error {
		events, err := app.FindCollectionByNameOrId("events")
		if err == nil {
			events.RemoveIndex("idx_events_slug")
			if err := app.Save(events); err != nil {
				return err
			}
		}

		templates, err := app.FindCollectionByNameOrId("templates")
		if err == nil {
			templates.RemoveIndex("idx_templates_slug")
			if err := app.Save(templates); err != nil {
				return err
			}
		}

		return nil
	})
}

func ensureUniqueSlug(app core.App, collection string) error {
	records, err := app.FindRecordsByFilter(collection, "slug != ''", "", 0, 0)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		slug := strings.TrimSpace(record.GetString("slug"))
		if slug == "" {
			continue
		}
		if _, exists := seen[slug]; exists {
			return fmt.Errorf("duplicate slug '%s' in %s; resolve before applying unique index", slug, collection)
		}
		seen[slug] = struct{}{}
	}

	return nil
}
