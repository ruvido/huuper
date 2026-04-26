package requests

import (
	"fmt"
	"net/url"
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const onboardingTokenService = "onboarding"

func BuildOnboardingURL(app *pocketbase.PocketBase, token string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	if base == "" || strings.TrimSpace(token) == "" {
		return ""
	}
	return base + "/onboarding/?token=" + url.QueryEscape(strings.TrimSpace(token))
}

func GenerateOnboardingToken(app *pocketbase.PocketBase, userID string) (string, error) {
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

func OnboardingUserForToken(app *pocketbase.PocketBase, token string) (*core.Record, *core.Record, error) {
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

	return tokenRecord, user, nil
}

func DeleteOnboardingToken(app *pocketbase.PocketBase, token string) error {
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
	if err != nil || tokenRecord == nil {
		return nil
	}

	return app.Delete(tokenRecord)
}

func DeleteOnboardingTokensForUser(app *pocketbase.PocketBase, userID string) {
	if app == nil || strings.TrimSpace(userID) == "" {
		return
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
		return
	}

	for _, record := range records {
		if record == nil {
			continue
		}
		_ = app.Delete(record)
	}
}
