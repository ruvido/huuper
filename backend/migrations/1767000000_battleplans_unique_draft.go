package migrations

import (
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		bp, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}

		// Drop pre-existing duplicate drafts (keep newest per user) so the
		// partial unique index below can be created on legacy data.
		drafts, err := app.FindRecordsByFilter(
			"battleplans",
			"status = 'draft'",
			"-updated",
			0, 0,
			map[string]any{},
		)
		if err != nil {
			return err
		}
		seen := map[string]struct{}{}
		for _, rec := range drafts {
			user := strings.TrimSpace(rec.GetString("user"))
			if user == "" {
				continue
			}
			if _, ok := seen[user]; ok {
				if err := app.Delete(rec); err != nil {
					return err
				}
				continue
			}
			seen[user] = struct{}{}
		}

		for _, idx := range bp.Indexes {
			if strings.Contains(idx, "idx_battleplans_user_draft") {
				return nil
			}
		}
		bp.AddIndex("idx_battleplans_user_draft", true, "user", "status = 'draft'")
		return app.Save(bp)
	}, func(app core.App) error {
		bp, err := app.FindCollectionByNameOrId("battleplans")
		if err != nil {
			return err
		}
		bp.RemoveIndex("idx_battleplans_user_draft")
		return app.Save(bp)
	})
}
