package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// GenerateTelegramTokenHandler creates a new token for Telegram connection
func GenerateTelegramTokenHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get authenticated user
		authRecord := e.Auth
		if authRecord == nil {
			return apis.NewUnauthorizedError("Unauthorized", nil)
		}

		// Generate cryptographically secure random token
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			return apis.NewBadRequestError("Failed to generate token", err)
		}
		token := hex.EncodeToString(bytes)

		// Get tokens collection
		tokensCollection, err := app.FindCollectionByNameOrId("tokens")
		if err != nil {
			return apis.NewNotFoundError("Tokens collection not found", err)
		}

		service := "telegram_connect"
		expiresAt := types.NowDateTime().Add(24 * time.Hour)

		// Invalidate previous tokens for this user/service
		oldTokens, err := app.FindRecordsByFilter(
			"tokens",
			"user = {:user} && service = {:service}",
			"",
			0,
			0,
			map[string]any{
				"user":    authRecord.Id,
				"service": service,
			},
		)
		if err == nil {
			for _, oldToken := range oldTokens {
				app.Delete(oldToken)
			}
		}

		// Create token record
		tokenRecord := core.NewRecord(tokensCollection)
		tokenRecord.Set("token", token)
		tokenRecord.Set("user", authRecord.Id)
		tokenRecord.Set("service", service)
		tokenRecord.Set("expires_at", expiresAt)

		if err := app.Save(tokenRecord); err != nil {
			return apis.NewBadRequestError("Failed to save token", err)
		}

		// Clean up expired tokens
		cleanupExpiredTokens(app)

		return e.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})
	}
}

// cleanupExpiredTokens removes expired tokens
func cleanupExpiredTokens(app *pocketbase.PocketBase) {
	now := types.NowDateTime()

	records, err := app.FindRecordsByFilter(
		"tokens",
		"expires_at < {:now}",
		"-expires_at",
		0,
		0,
		map[string]any{
			"now": now,
		},
	)

	if err != nil {
		return
	}

	for _, record := range records {
		app.Delete(record)
	}
}
