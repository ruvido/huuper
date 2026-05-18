package migrations

import (
	"fmt"
	"strings"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return resetOnboardingLimboUsers(app)
	}, func(app core.App) error {
		return nil
	})
}

// resetOnboardingLimboUsers detects users that were left in an
// inconsistent onboarding state by the previous flow (password set
// and/or completion marker set, but required profile fields still
// empty) and rolls them back to "pre-onboarding": password regenerated
// to a random unknown value, completion markers cleared, existing
// onboarding tokens removed, and a fresh onboarding token issued. The
// new onboarding URL is logged per user so the operator can re-send
// the onboarding email manually. Idempotent: running twice is a no-op
// because cleaned users no longer match the limbo criteria.
func resetOnboardingLimboUsers(app core.App) error {
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
		passwordSet := strings.TrimSpace(backendinternal.AnyToString(data["onboarding_password_set_at"])) != ""
		completionSet := strings.TrimSpace(backendinternal.AnyToString(data["onboarding_completed_at"])) != ""
		if !passwordSet && !completionSet {
			continue
		}
		if len(backendrequests.MissingOnboardingFields(data, settings, user)) == 0 {
			continue
		}

		onboardingURL, err := resetOnboardingLimboUser(app, user, data)
		if err != nil {
			return err
		}
		app.Logger().Info(
			"[migration limbo-onboarding] reset user",
			"user_id", user.Id,
			"email", strings.TrimSpace(user.GetString("email")),
			"onboarding_url", onboardingURL,
		)
	}

	return nil
}

func resetOnboardingLimboUser(app core.App, user *core.Record, data map[string]any) (string, error) {
	if user == nil {
		return "", fmt.Errorf("missing onboarding limbo user")
	}

	userID := user.Id
	var onboardingURL string
	err := app.RunInTransaction(func(txApp core.App) error {
		txUser, err := txApp.FindRecordById("users", userID)
		if err != nil {
			return err
		}

		resetData := make(map[string]any, len(data))
		for key, value := range data {
			resetData[key] = value
		}
		delete(resetData, "onboarding_completed_at")
		delete(resetData, "onboarding_password_set_at")
		newPwd := backendinternal.RandomToken()
		if newPwd == "" {
			return fmt.Errorf("failed to generate onboarding password for user %s", userID)
		}
		txUser.Set("password", newPwd)
		txUser.Set("passwordConfirm", newPwd)
		txUser.Set("data", resetData)

		if err := txApp.Save(txUser); err != nil {
			return err
		}

		if err := backendrequests.DeleteOnboardingTokensForUser(txApp, userID); err != nil {
			return err
		}

		token, err := backendrequests.GenerateOnboardingToken(txApp, userID)
		if err != nil {
			return err
		}
		url := backendrequests.BuildOnboardingURL(txApp, token)
		if strings.TrimSpace(url) == "" {
			return fmt.Errorf("invalid onboarding url for user %s", userID)
		}

		onboardingURL = url
		return nil
	})
	return onboardingURL, err
}
