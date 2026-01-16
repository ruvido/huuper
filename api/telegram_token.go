package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
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

		// Create token record
		tokenRecord := core.NewRecord(tokensCollection)
		tokenRecord.Set("token", token)
		tokenRecord.Set("user", authRecord.Id)
		tokenRecord.Set("service", "telegram")

		if err := app.Save(tokenRecord); err != nil {
			return apis.NewBadRequestError("Failed to save token", err)
		}

		// Clean up expired tokens (older than 24 hours)
		go cleanupExpiredTokens(app)

		return e.JSON(http.StatusOK, map[string]interface{}{
			"token": token,
		})
	}
}

// cleanupExpiredTokens removes tokens older than 24 hours
func cleanupExpiredTokens(app *pocketbase.PocketBase) {
	cutoff := time.Now().Add(-24 * time.Hour)

	records, err := app.FindRecordsByFilter(
		"tokens",
		"created < {:cutoff}",
		"-created",
		0,
		0,
		map[string]any{
			"cutoff": cutoff,
		},
	)

	if err != nil {
		return
	}

	for _, record := range records {
		app.Delete(record)
	}
}
