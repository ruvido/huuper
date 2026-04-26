package telegram

import (
	"log"
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

// RevokeAndDeleteUserInviteTokens revokes every outstanding Telegram invite
// link previously issued to the user and deletes the matching token records.
// Best-effort: a single failing group (missing chat_id, revoke error, ...)
// does not abort the loop so the caller can proceed with the broader cleanup.
func RevokeAndDeleteUserInviteTokens(app *pocketbase.PocketBase, userID string) {
	userID = strings.TrimSpace(userID)
	if app == nil || userID == "" {
		return
	}

	tokens, err := app.FindRecordsByFilter(
		"tokens",
		"user = {:user} && service = {:service}",
		"",
		0,
		0,
		map[string]any{
			"user":    userID,
			"service": inviteTokenService,
		},
	)
	if err != nil {
		log.Printf("[cancel] user=%s failed to load invite tokens: %v", userID, err)
		return
	}

	for _, token := range tokens {
		if token == nil {
			continue
		}
		groupID := strings.TrimSpace(token.GetString("group"))
		link := strings.TrimSpace(token.GetString("token"))

		if groupID != "" && link != "" {
			if group, err := app.FindRecordById("groups", groupID); err == nil && group != nil {
				if chatID, chatErr := TelegramChatIDForGroup(group); chatErr == nil {
					if revokeErr := RevokeInviteLink(chatID, link); revokeErr != nil {
						log.Printf("[cancel] user=%s group=%s revoke failed (non-blocking): %v", userID, groupID, revokeErr)
					}
				} else {
					log.Printf("[cancel] user=%s group=%s missing chat_id, skipping revoke: %v", userID, groupID, chatErr)
				}
			} else {
				log.Printf("[cancel] user=%s group=%s group lookup failed, skipping revoke: %v", userID, groupID, err)
			}
		}

		if delErr := app.Delete(token); delErr != nil {
			log.Printf("[cancel] user=%s group=%s failed to delete invite token: %v", userID, groupID, delErr)
		}
	}
}

// DeleteUserTelegramAuthTokens removes any pending telegram-binding tokens for
// the user. Separate from invite tokens — these are the short-lived codes used
// to associate a Telegram account with a webapp user.
func DeleteUserTelegramAuthTokens(app *pocketbase.PocketBase, userID string) {
	userID = strings.TrimSpace(userID)
	if app == nil || userID == "" {
		return
	}
	if err := deleteUserTokensByService(app, userID, "telegram", ""); err != nil {
		log.Printf("[cancel] user=%s failed to delete telegram auth tokens: %v", userID, err)
	}
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
