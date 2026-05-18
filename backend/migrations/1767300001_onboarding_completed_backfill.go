package migrations

import (
	"strings"
	"time"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return backfillOnboardingCompletion(app)
	}, func(app core.App) error {
		return nil
	})
}

func backfillOnboardingCompletion(app core.App) error {
	settings, err := backendrequests.LoadOnboardingSettings(app)
	if err != nil {
		return err
	}
	if len(settings.Steps) == 0 {
		return nil
	}

	users, err := app.FindRecordsByFilter("users", "", "", 0, 0)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user == nil {
			continue
		}

		data := backendinternal.ParseJSONMap(user.Get("data"))
		if strings.TrimSpace(backendinternal.AnyToString(data["onboarding_completed_at"])) != "" {
			continue
		}

		missing := backendrequests.MissingOnboardingFields(data, settings, user)
		if len(missing) > 0 {
			app.Logger().Info(
				"[migration onboarding-backfill] user requires onboarding",
				"user_id", user.Id,
				"email", strings.TrimSpace(user.GetString("email")),
				"missing", strings.Join(missing, ","),
			)
			continue
		}

		completedAt := strings.TrimSpace(user.GetString("created"))
		if completedAt == "" {
			completedAt = time.Now().UTC().Format(time.RFC3339)
		}
		data["onboarding_completed_at"] = completedAt
		user.Set("data", data)
		if err := app.Save(user); err != nil {
			return err
		}
		app.Logger().Info(
			"[migration onboarding-backfill] marked complete",
			"user_id", user.Id,
			"email", strings.TrimSpace(user.GetString("email")),
			"onboarding_completed_at", completedAt,
		)
	}

	return nil
}
