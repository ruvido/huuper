package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// VerifyTelegramAuth verifies the Telegram Login Widget signature
func VerifyTelegramAuth(data TelegramAuthData, botToken string) bool {
	// Check auth_date is not too old (within 24 hours)
	now := time.Now().Unix()
	if now-data.AuthDate > 86400 {
		return false
	}

	// Create data-check-string
	var dataCheckItems []string
	if data.ID != 0 {
		dataCheckItems = append(dataCheckItems, fmt.Sprintf("id=%d", data.ID))
	}
	if data.FirstName != "" {
		dataCheckItems = append(dataCheckItems, fmt.Sprintf("first_name=%s", data.FirstName))
	}
	if data.LastName != "" {
		dataCheckItems = append(dataCheckItems, fmt.Sprintf("last_name=%s", data.LastName))
	}
	if data.Username != "" {
		dataCheckItems = append(dataCheckItems, fmt.Sprintf("username=%s", data.Username))
	}
	if data.PhotoURL != "" {
		dataCheckItems = append(dataCheckItems, fmt.Sprintf("photo_url=%s", data.PhotoURL))
	}
	dataCheckItems = append(dataCheckItems, fmt.Sprintf("auth_date=%d", data.AuthDate))

	// Sort alphabetically
	sort.Strings(dataCheckItems)
	dataCheckString := strings.Join(dataCheckItems, "\n")

	// Calculate secret key
	h := sha256.New()
	h.Write([]byte(botToken))
	secretKey := h.Sum(nil)

	// Calculate hash
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(mac.Sum(nil))

	return calculatedHash == data.Hash
}

// LinkTelegramHandler handles linking a Telegram account to a user
func LinkTelegramHandler(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Get authenticated user (RequireAuth middleware ensures this is not nil)
		authRecord := e.Auth

		// Parse Telegram data
		var telegramData TelegramAuthData
		if err := e.BindBody(&telegramData); err != nil {
			return e.BadRequestError("Invalid request body", err)
		}

		// Verify Telegram signature
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		if botToken == "" {
			return e.InternalServerError("Bot token not configured", nil)
		}

		if !VerifyTelegramAuth(telegramData, botToken) {
			return e.ForbiddenError("Invalid Telegram signature", nil)
		}

		// Prepare telegram data to save
		telegramJSON := map[string]interface{}{
			"id":         strconv.FormatInt(telegramData.ID, 10),
			"first_name": telegramData.FirstName,
			"last_name":  telegramData.LastName,
			"username":   telegramData.Username,
			"photo_url":  telegramData.PhotoURL,
			"auth_date":  telegramData.AuthDate,
		}

		// Update user record
		authRecord.Set("telegram", telegramJSON)
		if err := app.Save(authRecord); err != nil {
			return e.InternalServerError("Failed to save telegram data", err)
		}

		// Return success
		return e.JSON(http.StatusOK, map[string]interface{}{
			"success":  true,
			"telegram": telegramJSON,
		})
	}
}
