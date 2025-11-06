package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase"
)

var bot *tgbotapi.BotAPI
var app *pocketbase.PocketBase

// StartTelegramBot initializes and starts the Telegram bot
func StartTelegramBot(pbApp *pocketbase.PocketBase) error {
	app = pbApp

	// Get bot token from settings
	telegramRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'telegram'",
		map[string]any{},
	)

	if err != nil {
		return fmt.Errorf("telegram settings not found: %w", err)
	}

	var telegramData struct {
		Token string `json:"token"`
		Name  string `json:"name"`
	}
	if err := telegramRecord.UnmarshalJSONField("data", &telegramData); err != nil {
		return fmt.Errorf("failed to parse telegram settings: %w", err)
	}

	if telegramData.Token == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	bot, err = tgbotapi.NewBotAPI(telegramData.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Telegram bot authorized: @%s", bot.Self.UserName)

	// Start listening for updates
	go listenForUpdates()

	return nil
}

func listenForUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Handle /start command with token
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			args := update.Message.CommandArguments()
			if args != "" {
				handleStartCommand(update.Message, args)
			}
		}
	}
}

func handleStartCommand(message *tgbotapi.Message, token string) {
	// Find token in database
	tokenRecord, err := app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = 'telegram'",
		map[string]any{
			"token": token,
		},
	)

	if err != nil {
		log.Printf("Invalid token: %s", token)
		reply := tgbotapi.NewMessage(message.Chat.ID, "❌ Invalid or expired token. Please try again from the dashboard.")
		bot.Send(reply)
		return
	}

	// Get user from token
	userId := tokenRecord.GetString("user")
	user, err := app.FindRecordById("users", userId)
	if err != nil {
		log.Printf("User not found: %s", userId)
		reply := tgbotapi.NewMessage(message.Chat.ID, "❌ User not found. Please try again.")
		bot.Send(reply)
		return
	}

	// Prepare Telegram data
	telegramData := map[string]interface{}{
		"id":         message.From.ID,
		"username":   message.From.UserName,
		"first_name": message.From.FirstName,
		"last_name":  message.From.LastName,
	}

	// Update user's telegram field (PocketBase handles JSON serialization)
	user.Set("telegram", telegramData)
	if err := app.Save(user); err != nil {
		log.Printf("Failed to update user: %v", err)
		reply := tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to save connection.")
		bot.Send(reply)
		return
	}

	// Delete used token
	if err := app.Delete(tokenRecord); err != nil {
		log.Printf("Failed to delete token: %v", err)
	}

	// Build success message
	email := user.GetString("email")
	username := message.From.UserName
	if username == "" {
		username = fmt.Sprintf("%s %s", message.From.FirstName, message.From.LastName)
	} else {
		username = "@" + username
	}

	// Get URL from settings
	urlRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'url'",
		map[string]any{},
	)

	url := "http://localhost:8090"
	if err == nil {
		var urlData struct {
			Address string `json:"address"`
		}
		if err := urlRecord.UnmarshalJSONField("data", &urlData); err == nil && urlData.Address != "" {
			url = urlData.Address
		}
	}

	profileURL := strings.TrimSuffix(url, "/") + "/#/profile"

	successMsg := fmt.Sprintf(
		"✅ Connected!\n\nEmail: %s\nTelegram: %s\n\nYou can close this chat or go back to the dashboard:\n%s",
		email,
		username,
		profileURL,
	)

	reply := tgbotapi.NewMessage(message.Chat.ID, successMsg)
	bot.Send(reply)

	log.Printf("Successfully connected user %s with Telegram %s", email, username)
}
