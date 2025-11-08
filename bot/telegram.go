package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var bot *tgbotapi.BotAPI
var app *pocketbase.PocketBase

// GetBot returns the bot instance
func GetBot() *tgbotapi.BotAPI {
	return bot
}

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
	u.AllowedUpdates = []string{"message", "my_chat_member", "chat_member"}

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Handle bot added/removed from groups
		if update.MyChatMember != nil {
			handleChatMemberUpdate(update.MyChatMember)
			continue
		}

		// Handle user added/removed from groups
		if update.ChatMember != nil {
			handleUserChatMemberUpdate(update.ChatMember)
			continue
		}

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

	// Sync user group memberships
	go syncUserGroupMemberships(user)

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

func handleChatMemberUpdate(update *tgbotapi.ChatMemberUpdated) {
	// Only handle groups/supergroups, not private chats
	if update.Chat.Type != "group" && update.Chat.Type != "supergroup" {
		return
	}

	newStatus := update.NewChatMember.Status
	chatID := update.Chat.ID

	log.Printf("Bot status changed in group '%s' (ID: %d): %s -> %s",
		update.Chat.Title, chatID, update.OldChatMember.Status, newStatus)

	// Bot became admin
	if newStatus == "administrator" {
		// Find existing group or create new
		chatIDStr := fmt.Sprintf("%d", chatID)
		group, err := app.FindFirstRecordByFilter(
			"groups",
			"telegram.chat_id = {:id}",
			map[string]any{"id": chatIDStr},
		)

		if err != nil {
			// Create new group
			collection, err := app.FindCollectionByNameOrId("groups")
			if err != nil {
				log.Printf("Failed to find groups collection: %v", err)
				return
			}
			group = core.NewRecord(collection)
		}

		// Update group data
		group.Set("name", update.Chat.Title)
		group.Set("type", "telegram")
		group.Set("telegram", map[string]any{
			"chat_id": chatIDStr,
			"type":    update.Chat.Type,
		})

		if err := app.Save(group); err != nil {
			log.Printf("Failed to save group: %v", err)
			return
		}

		log.Printf("Group '%s' saved successfully", update.Chat.Title)

		// Sync all connected users with new group
		go syncAllUsersWithNewGroup()
	}

	// Bot lost admin or was removed (member -> not admin, or kicked/left)
	if newStatus == "member" || newStatus == "left" || newStatus == "kicked" {
		chatIDStr := fmt.Sprintf("%d", chatID)
		group, err := app.FindFirstRecordByFilter(
			"groups",
			"telegram.chat_id = {:id}",
			map[string]any{"id": chatIDStr},
		)

		if err == nil && group != nil {
			// Delete all user_groups records first
			userGroupRecords, err := app.FindRecordsByFilter(
				"user_groups",
				"group = {:group}",
				"",
				0,
				0,
				map[string]any{"group": group.Id},
			)
			if err == nil {
				for _, ug := range userGroupRecords {
					app.Delete(ug)
				}
			}

			// Now delete the group
			if err := app.Delete(group); err != nil {
				log.Printf("Failed to delete group: %v", err)
				return
			}
			log.Printf("Group '%s' removed from database", update.Chat.Title)
		}
	}
}

func handleUserChatMemberUpdate(update *tgbotapi.ChatMemberUpdated) {
	// Only handle groups/supergroups
	if update.Chat.Type != "group" && update.Chat.Type != "supergroup" {
		return
	}

	userTelegramID := update.NewChatMember.User.ID
	newStatus := update.NewChatMember.Status
	chatID := update.Chat.ID

	// Find user by telegram ID
	user, err := app.FindFirstRecordByFilter(
		"users",
		"telegram.id = {:id}",
		map[string]any{"id": userTelegramID},
	)

	if err != nil {
		log.Printf("User with Telegram ID %d not found in DB", userTelegramID)
		return
	}

	// Find group by chat_id
	chatIDStr := fmt.Sprintf("%d", chatID)
	group, err := app.FindFirstRecordByFilter(
		"groups",
		"telegram.chat_id = {:id}",
		map[string]any{"id": chatIDStr},
	)

	if err != nil {
		log.Printf("Group with chat_id %d not found in DB", chatID)
		return
	}

	// User joined or became admin/creator
	if newStatus == "member" || newStatus == "administrator" || newStatus == "creator" {
		role := "member"
		if newStatus == "administrator" || newStatus == "creator" {
			role = "admin"
		}

		// Check if user_groups record exists
		existingRecord, _ := app.FindFirstRecordByFilter(
			"user_groups",
			"user = {:user} && group = {:group}",
			map[string]any{
				"user":  user.Id,
				"group": group.Id,
			},
		)

		if existingRecord != nil {
			// Update role if changed
			if existingRecord.GetString("role") != role {
				existingRecord.Set("role", role)
				if err := app.Save(existingRecord); err != nil {
					log.Printf("Failed to update user_groups role: %v", err)
				} else {
					log.Printf("✓ Updated user %s role to '%s' in group '%s'", user.GetString("email"), role, group.GetString("name"))
				}
			}
		} else {
			// Create new user_groups record
			userGroupsCollection, err := app.FindCollectionByNameOrId("user_groups")
			if err != nil {
				log.Printf("Failed to find user_groups collection: %v", err)
				return
			}

			userGroupRecord := core.NewRecord(userGroupsCollection)
			userGroupRecord.Set("user", user.Id)
			userGroupRecord.Set("group", group.Id)
			userGroupRecord.Set("role", role)

			if err := app.Save(userGroupRecord); err != nil {
				log.Printf("Failed to create user_groups record: %v", err)
			} else {
				log.Printf("✓ Added user %s to group '%s' with role '%s'", user.GetString("email"), group.GetString("name"), role)
			}
		}
	}

	// User left or was kicked
	if newStatus == "left" || newStatus == "kicked" {
		existingRecord, err := app.FindFirstRecordByFilter(
			"user_groups",
			"user = {:user} && group = {:group}",
			map[string]any{
				"user":  user.Id,
				"group": group.Id,
			},
		)

		if err == nil && existingRecord != nil {
			if err := app.Delete(existingRecord); err != nil {
				log.Printf("Failed to delete user_groups record: %v", err)
			} else {
				log.Printf("✓ Removed user %s from group '%s'", user.GetString("email"), group.GetString("name"))
			}
		}
	}
}

func syncAllUsersWithNewGroup() {
	users, err := app.FindRecordsByFilter(
		"users",
		"telegram.id != null && telegram.id != ''",
		"",
		0,
		0,
	)
	if err != nil {
		return
	}

	for _, user := range users {
		syncUserGroupMemberships(user)
	}
}

func syncUserGroupMemberships(user *core.Record) {
	// Get user's telegram data
	var telegramData struct {
		ID int64 `json:"id"`
	}

	if err := user.UnmarshalJSONField("telegram", &telegramData); err != nil {
		return
	}

	if telegramData.ID == 0 {
		return
	}

	// Get all telegram groups
	groups, err := app.FindRecordsByFilter("groups", "type = 'telegram'", "-created", 0, 0)
	if err != nil {
		return
	}

	for _, group := range groups {
		var telegramGroupData struct {
			ChatID string `json:"chat_id"`
		}

		if err := group.UnmarshalJSONField("telegram", &telegramGroupData); err != nil {
			continue
		}

		if telegramGroupData.ChatID == "" {
			continue
		}

		var chatID int64
		fmt.Sscanf(telegramGroupData.ChatID, "%d", &chatID)

		chatMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatID,
				UserID: telegramData.ID,
			},
		})

		if err != nil {
			continue
		}

		if chatMember.Status == "member" || chatMember.Status == "administrator" || chatMember.Status == "creator" {
			role := "member"
			if chatMember.Status == "administrator" || chatMember.Status == "creator" {
				role = "admin"
			}

			existingRecord, _ := app.FindFirstRecordByFilter(
				"user_groups",
				"user = {:user} && group = {:group}",
				map[string]any{
					"user":  user.Id,
					"group": group.Id,
				},
			)

			if existingRecord == nil {
				userGroupsCollection, _ := app.FindCollectionByNameOrId("user_groups")
				if userGroupsCollection != nil {
					userGroupRecord := core.NewRecord(userGroupsCollection)
					userGroupRecord.Set("user", user.Id)
					userGroupRecord.Set("group", group.Id)
					userGroupRecord.Set("role", role)
					app.Save(userGroupRecord)
				}
			}
		}
	}
}
