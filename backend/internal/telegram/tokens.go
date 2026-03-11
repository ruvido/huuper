package telegram

import (
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func GenerateUserToken(app *pocketbase.PocketBase, authRecord *core.Record) (string, error) {
	if authRecord == nil {
		return "", apis.NewUnauthorizedError("Unauthorized", nil)
	}

	token := backendinternal.RandomToken()
	if token == "" {
		return "", apis.NewBadRequestError("Failed to generate token", nil)
	}

	tokensCollection, err := app.FindCollectionByNameOrId("tokens")
	if err != nil {
		return "", apis.NewNotFoundError("Tokens collection not found", err)
	}

	oldTokens, err := app.FindRecordsByFilter(
		"tokens",
		"user = {:user} && service = 'telegram'",
		"",
		0,
		0,
		map[string]any{"user": authRecord.Id},
	)
	if err == nil {
		for _, oldToken := range oldTokens {
			_ = app.Delete(oldToken)
		}
	}

	tokenRecord := core.NewRecord(tokensCollection)
	tokenRecord.Set("token", token)
	tokenRecord.Set("user", authRecord.Id)
	tokenRecord.Set("service", "telegram")

	if err := app.Save(tokenRecord); err != nil {
		return "", apis.NewBadRequestError("Failed to save token", err)
	}

	cleanupExpiredTokens(app)
	return token, nil
}

func cleanupExpiredTokens(app *pocketbase.PocketBase) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	records, err := app.FindRecordsByFilter(
		"tokens",
		"created < {:cutoff}",
		"-created",
		0,
		0,
		map[string]any{"cutoff": cutoff},
	)
	if err != nil {
		return
	}
	for _, record := range records {
		_ = app.Delete(record)
	}
}
