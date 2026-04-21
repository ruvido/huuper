package bot

import (
	"fmt"
	"log"
	"strings"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"
	tginternal "members/backend/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var bot *tgbotapi.BotAPI
var app *pocketbase.PocketBase

type MembershipSyncStats struct {
	UsersChecked  int `json:"users_checked"`
	GroupsChecked int `json:"groups_checked"`
	Created       int `json:"created"`
	Updated       int `json:"updated"`
	Errors        int `json:"errors"`
}

// GetBot returns the bot instance
func GetBot() *tgbotapi.BotAPI {
	return bot
}

func SendMessage(chatID int64, message string) error {
	if bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := bot.Send(msg)
	return err
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
	tginternal.SetBotProvider(func() *tgbotapi.BotAPI {
		return bot
	})

	// Start listening for updates
	go listenForUpdates()
	// Catch-up for group names when the app was offline.
	go syncGroupNames()

	return nil
}

// StopTelegramBot stops the update receiver if it is running.
func StopTelegramBot() {
	if bot == nil {
		return
	}

	bot.StopReceivingUpdates()
	log.Printf("Telegram bot stopped")
}

func listenForUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = []string{"message", "my_chat_member", "chat_member", "chat_join_request"}

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

		if update.ChatJoinRequest != nil {
			handleChatJoinRequest(update.ChatJoinRequest)
			continue
		}

		if update.Message == nil {
			continue
		}

		// Handle group name change
		if update.Message.NewChatTitle != "" {
			updateGroupName(update.Message.Chat.ID, update.Message.NewChatTitle)
			continue
		}

		// Handle /start command with token
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			args := update.Message.CommandArguments()
			if args == "" {
				sendWarningMessage(update.Message.Chat.ID)
			} else {
				handleStartCommand(update.Message, args)
			}
			continue
		}

		// Handle private messages (non-commands)
		if update.Message.Chat.IsPrivate() && !update.Message.IsCommand() {
			handlePrivateMessage(update.Message)
		}
	}
}

func syncGroupNames() {
	groups, err := app.FindRecordsByFilter("groups", "telegram.chat_id != ''", "", 0, 0)
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
		if chatID == 0 {
			continue
		}

		chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		})
		if err != nil || chat.Title == "" {
			continue
		}

		if group.GetString("name") == chat.Title {
			continue
		}

		group.Set("name", chat.Title)
		if err := app.Save(group); err != nil {
			log.Printf("Failed to sync group name: %v", err)
		} else {
			log.Printf("✓ Synced group name to '%s' (ID: %d)", chat.Title, chatID)
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

func handleChatJoinRequest(request *tgbotapi.ChatJoinRequest) {
	if request == nil || request.InviteLink == nil {
		return
	}

	inviteLink := strings.TrimSpace(request.InviteLink.InviteLink)
	if inviteLink == "" {
		return
	}

	user, err := tginternal.ConsumeMemberInvite(app, inviteLink)
	if err != nil {
		log.Printf("Failed to resolve invite link %q: %v", inviteLink, err)
		notifyJoinFailure("resolve invite link", request, inviteLink, nil, err)
		declineChatJoinRequest(request)
		return
	}

	telegramData := backendinternal.ParseJSONMap(user.Get("telegram"))
	telegramData["id"] = request.From.ID
	if username := strings.TrimSpace(request.From.UserName); username != "" {
		telegramData["username"] = username
	}
	if firstName := strings.TrimSpace(request.From.FirstName); firstName != "" {
		telegramData["first_name"] = firstName
	}
	if lastName := strings.TrimSpace(request.From.LastName); lastName != "" {
		telegramData["last_name"] = lastName
	}
	telegramData["invite_link"] = inviteLink
	user.Set("telegram", telegramData)
	if err := app.Save(user); err != nil {
		log.Printf("Failed to save Telegram data for user %s: %v", user.Id, err)
		notifyJoinFailure("save Telegram data", request, inviteLink, user, err)
		declineChatJoinRequest(request)
		return
	}

	if _, err := bot.Request(tgbotapi.ApproveChatJoinRequestConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: request.Chat.ID},
		UserID:     request.From.ID,
	}); err != nil {
		log.Printf("Failed to approve chat join request for user %d in chat %d: %v", request.From.ID, request.Chat.ID, err)
		notifyJoinFailure("approve chat join request", request, inviteLink, user, err)
		declineChatJoinRequest(request)
		return
	}

	if _, err := bot.Request(tgbotapi.RevokeChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: request.Chat.ID},
		InviteLink: inviteLink,
	}); err != nil {
		log.Printf("Failed to revoke invite link %q for chat %d: %v", inviteLink, request.Chat.ID, err)
	}

	syncUserGroupRecord(user, request.Chat.ID, "member")
	log.Printf("Approved chat join request for user %s via invite %q in chat %d", user.GetString("email"), inviteLink, request.Chat.ID)
}

func notifyJoinFailure(stage string, request *tgbotapi.ChatJoinRequest, inviteLink string, user *core.Record, cause error) {
	if app == nil {
		return
	}

	from := request.From

	userEmail := ""
	userName := ""
	telegramUsername := strings.TrimSpace(from.UserName)
	if user != nil {
		userEmail = strings.TrimSpace(user.GetString("email"))
		userData := backendinternal.ParseJSONMap(user.Get("data"))
		userName = strings.TrimSpace(backendinternal.AnyToString(userData["full_name"]))
	}
	if userName == "" {
		userName = strings.TrimSpace(strings.TrimSpace(from.FirstName + " " + from.LastName))
	}
	subjectTarget := userEmail
	if subjectTarget == "" {
		subjectTarget = telegramUsername
	}
	if subjectTarget == "" {
		subjectTarget = fmt.Sprintf("telegram:%d", from.ID)
	}

	subject := "Telegram invite process failed"
	if subjectTarget != "" {
		subject = "Telegram invite process failed for " + subjectTarget
	}

	body := strings.Join([]string{
		"Telegram invite process failed.",
		"",
		"Stage: " + strings.TrimSpace(stage),
		"Reason: " + strings.TrimSpace(cause.Error()),
		"User: " + strings.TrimSpace(userName),
		"Email: " + strings.TrimSpace(userEmail),
		"Telegram username: " + strings.TrimSpace(telegramUsername),
		fmt.Sprintf("Telegram ID: %d", from.ID),
		fmt.Sprintf("Chat ID: %d", request.Chat.ID),
		"Chat title: " + strings.TrimSpace(request.Chat.Title),
		"Invite link: " + strings.TrimSpace(inviteLink),
	}, "\n")

	if !backendinternal.SendAdminFailureEmail(app, subject, body) {
		log.Printf("Failed to send admin failure email for telegram invite process (%s) user=%s cause=%v", stage, userEmail, cause)
	}
}

func declineChatJoinRequest(request *tgbotapi.ChatJoinRequest) {
	if request == nil {
		return
	}
	if bot == nil {
		return
	}

	_, err := bot.Request(tgbotapi.DeclineChatJoinRequest{
		ChatConfig: tgbotapi.ChatConfig{ChatID: request.Chat.ID},
		UserID:     request.From.ID,
	})
	if err != nil {
		log.Printf("Failed to decline chat join request for user %d in chat %d: %v", request.From.ID, request.Chat.ID, err)
	}
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
		group.Set("type", "local")
		group.Set("telegram", map[string]any{
			"chat_id": chatIDStr,
			"type":    update.Chat.Type,
		})

		if err := app.Save(group); err != nil {
			log.Printf("Failed to save group: %v", err)
			return
		}

		log.Printf("Group '%s' saved successfully", update.Chat.Title)

		// Send welcome message
		go sendWelcomeMessage(chatID)

		// Sync all connected users with new group
		go SyncAllUsersMemberships()
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

	if newStatus == "member" || newStatus == "administrator" || newStatus == "creator" {
		user, err := app.FindFirstRecordByFilter(
			"users",
			"telegram.id = {:id}",
			map[string]any{"id": userTelegramID},
		)

		if err != nil {
			if err := removeUnregisteredTelegramUser(chatID, userTelegramID); err != nil {
				log.Printf("Failed to remove unregistered Telegram user %d from group '%s': %v", userTelegramID, update.Chat.Title, err)
			} else {
				log.Printf("Removed unregistered Telegram user %d from group '%s'", userTelegramID, update.Chat.Title)
			}
			return
		}

		syncUserGroupRecord(user, chatID, newStatus)
		return
	}

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

func syncUserGroupRecord(user *core.Record, chatID int64, status string) {
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
	if err := groupinternal.EnsureMembership(app, user.Id, group.Id); err != nil {
		log.Printf("Failed to create user_groups record: %v", err)
	} else {
		log.Printf("✓ Added user %s to group '%s'", user.GetString("email"), group.GetString("name"))
	}
}

func removeUnregisteredTelegramUser(chatID int64, userTelegramID int64) error {
	if bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	ban := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userTelegramID,
		},
		RevokeMessages: false,
	}
	if _, err := bot.Request(ban); err != nil {
		return err
	}

	unban := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userTelegramID,
		},
		OnlyIfBanned: true,
	}
	_, err := bot.Request(unban)
	return err
}

// SyncAllUsersMemberships re-checks Telegram memberships for all users
// with a connected Telegram account and restores missing user_groups records.
func SyncAllUsersMemberships() {
	_, _ = SyncAllUsersMembershipsWithStats()
}

// SyncAllUsersMembershipsWithStats re-checks Telegram memberships for all users
// with a connected Telegram account and restores missing user_groups records.
func SyncAllUsersMembershipsWithStats() (MembershipSyncStats, error) {
	stats := MembershipSyncStats{}
	if app == nil {
		return stats, fmt.Errorf("pocketbase app not initialized")
	}
	if bot == nil {
		return stats, fmt.Errorf("telegram bot not initialized")
	}

	users, err := app.FindRecordsByFilter(
		"users",
		"telegram.id != null && telegram.id != ''",
		"",
		0,
		0,
	)
	if err != nil {
		return stats, err
	}

	for _, user := range users {
		stats.UsersChecked++
		partial := syncUserGroupMemberships(user)
		stats.GroupsChecked += partial.GroupsChecked
		stats.Created += partial.Created
		stats.Updated += partial.Updated
		stats.Errors += partial.Errors
	}

	return stats, nil
}

func syncUserGroupMemberships(user *core.Record) MembershipSyncStats {
	stats := MembershipSyncStats{}

	// Get user's telegram data
	var telegramData struct {
		ID int64 `json:"id"`
	}

	if err := user.UnmarshalJSONField("telegram", &telegramData); err != nil {
		stats.Errors++
		return stats
	}

	if telegramData.ID == 0 {
		return stats
	}

	// Get all groups and process those that have a Telegram chat_id.
	groups, err := app.FindRecordsByFilter(
		"groups",
		"",
		"-created",
		0,
		0,
	)
	if err != nil {
		stats.Errors++
		return stats
	}

	for _, group := range groups {
		stats.GroupsChecked++

		var telegramGroupData struct {
			ChatID string `json:"chat_id"`
		}

		if err := group.UnmarshalJSONField("telegram", &telegramGroupData); err != nil {
			stats.Errors++
			continue
		}

		if telegramGroupData.ChatID == "" {
			continue
		}

		var chatID int64
		if _, err := fmt.Sscanf(telegramGroupData.ChatID, "%d", &chatID); err != nil || chatID == 0 {
			stats.Errors++
			continue
		}

		chatMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatID,
				UserID: telegramData.ID,
			},
		})

		if err != nil {
			stats.Errors++
			continue
		}

		if chatMember.Status == "member" || chatMember.Status == "administrator" || chatMember.Status == "creator" {
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
					if err := app.Save(userGroupRecord); err != nil {
						stats.Errors++
					} else {
						stats.Created++
					}
				} else {
					stats.Errors++
				}
			}
		}
	}

	return stats
}

func sendWelcomeMessage(chatID int64) {
	// Get URL from settings
	urlRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'url'",
		map[string]any{},
	)
	if err != nil {
		log.Printf("Failed to get URL settings: %v", err)
		return
	}

	var urlData struct {
		Address string `json:"address"`
	}
	if err := urlRecord.UnmarshalJSONField("data", &urlData); err != nil {
		log.Printf("Failed to parse URL settings: %v", err)
		return
	}

	// Get welcome message from settings
	messagesRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'bot_messages'",
		map[string]any{},
	)
	if err != nil {
		log.Printf("Failed to get bot messages settings: %v", err)
		return
	}

	var messagesData struct {
		Welcome string `json:"welcome"`
	}
	if err := messagesRecord.UnmarshalJSONField("data", &messagesData); err != nil {
		log.Printf("Failed to parse bot messages settings: %v", err)
		return
	}

	// Replace {url} placeholder
	message := strings.ReplaceAll(messagesData.Welcome, "{url}", urlData.Address)

	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
	} else {
		log.Printf("Sent welcome message to chat %d", chatID)
	}
}

func handlePrivateMessage(message *tgbotapi.Message) {
	sendWarningMessage(message.Chat.ID)
}

func sendWarningMessage(chatID int64) {
	// Get URL from settings
	urlRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'url'",
		map[string]any{},
	)
	if err != nil {
		log.Printf("Failed to get URL settings: %v", err)
		return
	}

	var urlData struct {
		Address string `json:"address"`
	}
	if err := urlRecord.UnmarshalJSONField("data", &urlData); err != nil {
		log.Printf("Failed to parse URL settings: %v", err)
		return
	}

	// Get warning message from settings
	messagesRecord, err := app.FindFirstRecordByFilter(
		"settings",
		"name = 'bot_messages'",
		map[string]any{},
	)
	if err != nil {
		log.Printf("Failed to get bot messages settings: %v", err)
		return
	}

	var messagesData struct {
		Warning string `json:"warning"`
	}
	if err := messagesRecord.UnmarshalJSONField("data", &messagesData); err != nil {
		log.Printf("Failed to parse bot messages settings: %v", err)
		return
	}

	message := strings.ReplaceAll(messagesData.Warning, "{url}", urlData.Address)
	if message == "" {
		return
	}

	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send warning message: %v", err)
	}
}

func updateGroupName(chatID int64, newTitle string) {
	chatIDStr := fmt.Sprintf("%d", chatID)

	group, err := app.FindFirstRecordByFilter(
		"groups",
		"telegram.chat_id = {:id}",
		map[string]any{"id": chatIDStr},
	)

	if err != nil {
		return
	}

	group.Set("name", newTitle)
	if err := app.Save(group); err != nil {
		log.Printf("Failed to update group name: %v", err)
	} else {
		log.Printf("✓ Updated group name to '%s' (ID: %d)", newTitle, chatID)
	}
}
