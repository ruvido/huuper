package hooks

import (
	"strings"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterUsersNormalization normalizes selected user profile fields.
func RegisterUsersNormalization(app *pocketbase.PocketBase) {
	app.OnRecordValidate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := normalizeUserEmailAndValidateUnique(e.App, e.Record); err != nil {
			return err
		}
		if err := normalizeUserProfileData(e.Record); err != nil {
			return err
		}
		return e.Next()
	})
}

// RegisterUsersAuthGate blocks authentication for users that haven't
// finished the onboarding flow yet. This is defense-in-depth: under
// normal flow such users have a random unknown password and cannot
// log in via password, but the hook also blocks OAuth2 and any path
// that bypasses password verification, so it is impossible to reach
// the API as an authenticated user with an empty profile.
func RegisterUsersAuthGate(app *pocketbase.PocketBase) {
	app.OnRecordAuthRequest("users").BindFunc(func(e *core.RecordAuthRequestEvent) error {
		if e.Record == nil {
			return e.Next()
		}
		data := backendinternal.ParseJSONMap(e.Record.Get("data"))
		if strings.TrimSpace(backendinternal.AnyToString(data["onboarding_completed_at"])) == "" {
			token, err := backendrequests.EnsureOnboardingToken(e.App, e.Record.Id)
			if err != nil {
				return apis.NewForbiddenError("onboarding_incomplete", map[string]any{"error": err.Error()})
			}
			return apis.NewForbiddenError("onboarding_incomplete", map[string]any{
				"onboarding_url": backendrequests.BuildRelativeOnboardingURL(token),
			})
		}
		return e.Next()
	})
}

func normalizeUserEmailAndValidateUnique(app core.App, record *core.Record) error {
	if record == nil {
		return nil
	}

	email := strings.ToLower(strings.TrimSpace(record.GetString("email")))
	if email == "" {
		return nil
	}
	record.Set("email", email)

	if original := record.Original(); original != nil {
		previous := strings.ToLower(strings.TrimSpace(original.GetString("email")))
		if previous == email {
			return nil
		}
	}

	records, err := app.FindRecordsByFilter("users", "", "", 0, 0)
	if err != nil {
		return apis.NewBadRequestError("failed_to_validate_email_uniqueness", err)
	}

	for _, existing := range records {
		if existing == nil || existing.Id == record.Id {
			continue
		}
		existingEmail := strings.TrimSpace(existing.GetString("email"))
		if existingEmail == "" {
			continue
		}
		if strings.EqualFold(existingEmail, email) {
			return apis.NewBadRequestError("email_already_exists", nil)
		}
	}

	return nil
}

func normalizeUserProfileData(record *core.Record) error {
	if record == nil {
		return nil
	}

	data := backendinternal.ParseJSONMap(record.Get("data"))
	if mobile, ok := data["mobile"].(string); ok {
		normalized, err := backendinternal.NormalizePhone(mobile)
		if err != nil {
			return apis.NewBadRequestError("invalid_mobile", err)
		}
		data["mobile"] = normalized
	}

	value, ok := data["full_name"].(string)
	if !ok {
		record.Set("data", data)
		return nil
	}

	normalized := backendinternal.NormalizePersonName(value)
	if normalized == "" {
		record.Set("data", data)
		return nil
	}

	data["full_name"] = normalized
	record.Set("data", data)
	return nil
}
