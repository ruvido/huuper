package telegram

import (
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const inviteTokenService = "telegram_invite"

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

	if err := deleteUserTokensByService(app, authRecord.Id, "telegram", ""); err != nil {
		return "", err
	}

	if err := saveUserToken(app, tokensCollection, authRecord.Id, "", "telegram", token); err != nil {
		return "", err
	}

	cleanupExpiredTokens(app)
	return token, nil
}

func GenerateInviteToken(app *pocketbase.PocketBase, userID string, groupID string, inviteLink string) error {
	if strings.TrimSpace(userID) == "" {
		return apis.NewBadRequestError("missing_user", nil)
	}
	groupID = strings.TrimSpace(groupID)
	inviteLink = strings.TrimSpace(inviteLink)
	if inviteLink == "" {
		return apis.NewBadRequestError("missing_invite_link", nil)
	}

	tokensCollection, err := app.FindCollectionByNameOrId("tokens")
	if err != nil {
		return apis.NewNotFoundError("Tokens collection not found", err)
	}

	if err := deleteUserTokensByService(app, userID, inviteTokenService, groupID); err != nil {
		return err
	}

	if err := saveUserToken(app, tokensCollection, userID, groupID, inviteTokenService, inviteLink); err != nil {
		return err
	}

	cleanupExpiredTokens(app)
	return nil
}

func DeleteInviteToken(app core.App, userID string, groupID string) error {
	return deleteUserTokensByService(app, userID, inviteTokenService, groupID)
}

func deleteUserTokensByService(app core.App, userID string, service string, groupID string) error {
	filter := "user = {:user} && service = {:service}"
	params := map[string]any{
		"user":    userID,
		"service": service,
	}
	if strings.TrimSpace(groupID) != "" {
		filter += " && group = {:group}"
		params["group"] = strings.TrimSpace(groupID)
	}

	oldTokens, err := app.FindRecordsByFilter(
		"tokens",
		filter,
		"",
		0,
		0,
		params,
	)
	if err != nil {
		return err
	}
	for _, oldToken := range oldTokens {
		_ = app.Delete(oldToken)
	}
	return nil
}

func saveUserToken(app core.App, tokensCollection *core.Collection, userID string, groupID string, service string, token string) error {
	tokenRecord := core.NewRecord(tokensCollection)
	tokenRecord.Set("token", token)
	tokenRecord.Set("user", userID)
	if strings.TrimSpace(groupID) != "" {
		tokenRecord.Set("group", strings.TrimSpace(groupID))
	}
	tokenRecord.Set("service", service)
	if err := app.Save(tokenRecord); err != nil {
		return apis.NewBadRequestError("Failed to save token", err)
	}
	return nil
}

func cleanupExpiredTokens(app *pocketbase.PocketBase) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	records, err := app.FindRecordsByFilter(
		"tokens",
		"created < {:cutoff} && service = {:service}",
		"-created",
		0,
		0,
		map[string]any{"cutoff": cutoff, "service": "telegram"},
	)
	if err != nil {
		return
	}
	for _, record := range records {
		_ = app.Delete(record)
	}
}
