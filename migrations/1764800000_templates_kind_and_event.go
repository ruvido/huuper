package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	templateKindUserRegistrationReceived = "events.user.registration_received"
	templateKindAdminNewRegistration     = "events.admin.new_registration"
)

func init() {
	m.Register(func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		events, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}

		if templates.Fields.GetByName("kind") == nil {
			templates.Fields.Add(&core.TextField{
				Name:     "kind",
				Required: false,
				Max:      200,
			})
		}

		if templates.Fields.GetByName("event") == nil {
			templates.Fields.Add(&core.RelationField{
				Name:         "event",
				Required:     false,
				CollectionId: events.Id,
				MaxSelect:    1,
			})
		}

		templates.AddIndex("idx_templates_kind_event", true, "kind,event", "kind != ''")
		if err := app.Save(templates); err != nil {
			return err
		}

		if err := backfillAdminTemplateKind(app); err != nil {
			return err
		}

		return backfillEventReplyTemplateKinds(app)
	}, func(app core.App) error {
		templates, err := app.FindCollectionByNameOrId("templates")
		if err != nil {
			return err
		}

		templates.RemoveIndex("idx_templates_kind_event")

		if field := templates.Fields.GetByName("event"); field != nil {
			templates.Fields.RemoveById(field.GetId())
		}
		if field := templates.Fields.GetByName("kind"); field != nil {
			templates.Fields.RemoveById(field.GetId())
		}

		return app.Save(templates)
	})
}

func backfillAdminTemplateKind(app core.App) error {
	records, err := app.FindRecordsByFilter(
		"templates",
		"slug = {:slug}",
		"",
		0,
		0,
		map[string]any{"slug": "admin-email-event"},
	)
	if err != nil {
		return err
	}

	for _, record := range records {
		if strings.TrimSpace(record.GetString("kind")) != "" {
			continue
		}
		record.Set("kind", templateKindAdminNewRegistration)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}

func backfillEventReplyTemplateKinds(app core.App) error {
	events, err := app.FindRecordsByFilter("events", "reply_template != ''", "", 0, 0)
	if err != nil {
		return err
	}

	for _, event := range events {
		templateID := strings.TrimSpace(event.GetString("reply_template"))
		if templateID == "" {
			continue
		}

		template, err := app.FindRecordById("templates", templateID)
		if err != nil || template == nil {
			continue
		}

		changed := false
		if strings.TrimSpace(template.GetString("kind")) == "" {
			template.Set("kind", templateKindUserRegistrationReceived)
			changed = true
		}
		if strings.TrimSpace(template.GetString("event")) == "" {
			template.Set("event", event.Id)
			changed = true
		}

		if changed {
			if err := app.Save(template); err != nil {
				return err
			}
		}
	}

	return nil
}
