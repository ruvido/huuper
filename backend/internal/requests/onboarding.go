package requests

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const onboardingTokenService = "onboarding"

func BuildOnboardingURL(app core.App, token string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	if base == "" || strings.TrimSpace(token) == "" {
		return ""
	}
	return base + BuildRelativeOnboardingURL(token)
}

func BuildRelativeOnboardingURL(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	return "/onboarding/?token=" + url.QueryEscape(trimmed)
}

// MissingOnboardingFields returns the field names from the onboarding
// settings that are not populated (empty string / empty array / nil) in
// the provided user data. An empty result means the user data satisfies
// all required onboarding steps.
func MissingOnboardingFields(data map[string]any, settings OnboardingSettingsConfig, user *core.Record) []string {
	if len(settings.Steps) == 0 {
		return nil
	}
	missing := []string{}
	for _, step := range settings.Steps {
		field := strings.TrimSpace(step.Field)
		if field == "" {
			continue
		}
		if !hasNonEmptyOnboardingValue(data[field]) {
			if user == nil || user.Collection() == nil || user.Collection().Fields.GetByName(field) == nil || !hasNonEmptyOnboardingValue(user.Get(field)) {
				missing = append(missing, field)
			}
		}
	}
	return missing
}

func hasNonEmptyOnboardingValue(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case []string:
		return len(typed) > 0
	case []any:
		return len(typed) > 0
	default:
		return true
	}
}

func OnboardingCompleteForUser(app core.App, user *core.Record) bool {
	if user == nil {
		return false
	}

	data := backendinternal.ParseJSONMap(user.Get("data"))
	return strings.TrimSpace(backendinternal.AnyToString(data["onboarding_completed_at"])) != ""
}

func GenerateOnboardingToken(app core.App, userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", apis.NewBadRequestError("missing_user", nil)
	}

	tokensCollection, err := app.FindCollectionByNameOrId("tokens")
	if err != nil {
		return "", apis.NewNotFoundError("tokens_collection_not_found", err)
	}

	for i := 0; i < 5; i++ {
		token := backendinternal.RandomToken()
		if token == "" {
			continue
		}

		tokenRecord := core.NewRecord(tokensCollection)
		tokenRecord.Set("token", token)
		tokenRecord.Set("user", userID)
		tokenRecord.Set("service", onboardingTokenService)
		if err := app.Save(tokenRecord); err == nil {
			return token, nil
		}
	}

	return "", fmt.Errorf("unable to generate onboarding token")
}

func EnsureOnboardingToken(app core.App, userID string) (string, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return "", apis.NewBadRequestError("missing_user", nil)
	}

	tokenRecord, err := app.FindFirstRecordByFilter(
		"tokens",
		"user = {:user} && service = {:service}",
		map[string]any{
			"user":    trimmedUserID,
			"service": onboardingTokenService,
		},
	)
	if err == nil && tokenRecord != nil {
		token := strings.TrimSpace(tokenRecord.GetString("token"))
		if token != "" {
			return token, nil
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	return GenerateOnboardingToken(app, trimmedUserID)
}

func OnboardingUserForToken(app core.App, token string) (*core.Record, *core.Record, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil, nil, apis.NewBadRequestError("missing_token", nil)
	}

	tokenRecord, err := app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = {:service}",
		map[string]any{
			"token":   trimmed,
			"service": onboardingTokenService,
		},
	)
	if err != nil || tokenRecord == nil {
		return nil, nil, apis.NewBadRequestError("invalid_token", err)
	}

	userID := strings.TrimSpace(tokenRecord.GetString("user"))
	if userID == "" {
		return nil, nil, apis.NewBadRequestError("invalid_token", nil)
	}

	user, err := app.FindRecordById("users", userID)
	if err != nil || user == nil {
		return nil, nil, apis.NewBadRequestError("invalid_token", err)
	}
	if OnboardingCompleteForUser(app, user) {
		return nil, nil, apis.NewBadRequestError("invalid_token", nil)
	}

	return tokenRecord, user, nil
}

func DeleteOnboardingToken(app core.App, token string) error {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil
	}

	tokenRecord, err := app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = {:service}",
		map[string]any{
			"token":   trimmed,
			"service": onboardingTokenService,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if tokenRecord == nil {
		return nil
	}

	return app.Delete(tokenRecord)
}

func DeleteOnboardingTokensForUser(app core.App, userID string) error {
	if app == nil || strings.TrimSpace(userID) == "" {
		return nil
	}

	records, err := app.FindRecordsByFilter(
		"tokens",
		"user = {:user} && service = {:service}",
		"",
		500,
		0,
		map[string]any{
			"user":    strings.TrimSpace(userID),
			"service": onboardingTokenService,
		},
	)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record == nil {
			continue
		}
		if err := app.Delete(record); err != nil {
			return err
		}
	}
	return nil
}
