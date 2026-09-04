package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// retreats is a standalone module for rare, public, deposit-gated
// gatherings ("raduni"). It is intentionally independent from `events`
// (call/meetup engine): no recurrence, no per-type permission matrix, no
// public/me/admin route triplication. See docs/CLAUDE.md and the approved
// plan for the full rationale.
func init() {
	m.Register(func(app core.App) error {
		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		retreats := core.NewBaseCollection("retreats")
		retreats.ListRule = types.Pointer("") // Public read (detail page never gates on `active`)
		retreats.ViewRule = types.Pointer("") // Public read (also needed for public file access to `gallery`)
		retreats.CreateRule = nil             // Admin only
		retreats.UpdateRule = nil             // Admin only
		retreats.DeleteRule = nil             // Admin only

		retreats.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			},
			&core.TextField{
				Name:     "title",
				Required: true,
				Max:      200,
			},
			&core.TextField{
				Name:     "tagline",
				Required: false,
				Max:      200,
			},
			&core.TextField{
				Name:     "slug",
				Required: true,
				Max:      200,
			},
			&core.TextField{
				Name:     "location",
				Required: false,
				Max:      200,
			},
			&core.DateField{
				Name:     "start_date",
				Required: true,
			},
			&core.DateField{
				Name:     "end_date",
				Required: false,
			},
			&core.BoolField{
				Name:     "active",
				Required: false,
			},
			&core.RelationField{
				Name:         "telegram_group",
				Required:     false,
				CollectionId: groups.Id,
				MaxSelect:    1,
			},
			&core.FileField{
				Name:      "gallery",
				Required:  false,
				MaxSelect: 30,
				MaxSize:   15 * 1024 * 1024,
				MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
			},
			&core.JSONField{
				Name:     "data",
				Required: false,
			},
		)

		retreats.AddIndex("idx_retreats_slug", true, "slug", "")

		return app.Save(retreats)
	}, func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return nil
		}
		return app.Delete(retreats)
	})
}
