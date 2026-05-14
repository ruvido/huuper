package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var botProvider func() *tgbotapi.BotAPI

func SetBotProvider(provider func() *tgbotapi.BotAPI) {
	botProvider = provider
}

func currentBot() *tgbotapi.BotAPI {
	if botProvider == nil {
		return nil
	}
	return botProvider()
}

func DefaultGroupInvite(app *pocketbase.PocketBase, actor *core.Record) (string, string, error) {
	if actor == nil {
		return "", "", apis.NewUnauthorizedError("Unauthorized", nil)
	}
	if strings.TrimSpace(actor.GetString("status")) == "cancelled" {
		return "", "", apis.NewForbiddenError("user_cancelled", nil)
	}
	log.Printf("[default-invite] user=%s request started", strings.TrimSpace(actor.Id))

	group, err := app.FindFirstRecordByFilter("groups", "type = 'local'", map[string]any{})
	if err != nil || group == nil {
		return "", "", apis.NewNotFoundError("local_group_not_found", err)
	}

	telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
	chatIDRaw := strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"]))
	if chatIDRaw == "" {
		return "", "", apis.NewBadRequestError("invalid_default_group", fmt.Errorf("missing telegram.chat_id"))
	}

	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return "", "", apis.NewBadRequestError("invalid_default_group", fmt.Errorf("invalid telegram.chat_id"))
	}

	tg := currentBot()
	if tg == nil {
		return "", "", apis.NewBadRequestError("telegram_bot_unavailable", fmt.Errorf("telegram bot is not initialized"))
	}

	logMembershipDiagnostics(tg, actor, chatID)

	invite, err := createChatInviteLink(tg, chatID, "", false, time.Time{})
	if err != nil {
		return "", "", apis.NewBadRequestError("invite_link_generation_failed", err)
	}

	link := strings.TrimSpace(invite.InviteLink)
	if link == "" {
		log.Printf("[default-invite] user=%s chat_id=%d empty link from createChatInviteLink", strings.TrimSpace(actor.Id), chatID)
		return "", "", apis.NewBadRequestError("invite_link_generation_failed", fmt.Errorf("empty invite link"))
	}
	nowUnix := time.Now().Unix()
	if invite.IsRevoked {
		return "", "", apis.NewBadRequestError("invite_link_generation_failed", fmt.Errorf("created invite link is revoked"))
	}
	if invite.ExpireDate > 0 && int64(invite.ExpireDate) <= nowUnix {
		return "", "", apis.NewBadRequestError("invite_link_generation_failed", fmt.Errorf("created invite link is expired"))
	}
	log.Printf("[default-invite] user=%s chat_id=%d created invite is_primary=%v is_revoked=%v expire_date=%d member_limit=%d creates_join_request=%v",
		strings.TrimSpace(actor.Id), chatID, invite.IsPrimary, invite.IsRevoked, invite.ExpireDate, invite.MemberLimit, invite.CreatesJoinRequest)

	return group.Id, link, nil
}

func ResolveMemberInviteForChat(app *pocketbase.PocketBase, inviteLink string, chatID int64) (*core.Record, string, error) {
	trimmed := strings.TrimSpace(inviteLink)
	if trimmed == "" {
		return nil, "", apis.NewBadRequestError("missing_invite_link", nil)
	}

	tokenRecord, err := inviteTokenForJoinRequest(app, trimmed, chatID)
	if err != nil || tokenRecord == nil {
		return nil, "", apis.NewBadRequestError("invalid_invite_link", err)
	}

	userID := strings.TrimSpace(tokenRecord.GetString("user"))
	if userID == "" {
		return nil, "", apis.NewBadRequestError("invalid_invite_link", nil)
	}

	resolved, err := app.FindRecordById("users", userID)
	if err != nil || resolved == nil {
		return nil, "", apis.NewBadRequestError("invalid_invite_link", err)
	}

	return resolved, strings.TrimSpace(tokenRecord.GetString("token")), nil
}

func inviteTokenForJoinRequest(app *pocketbase.PocketBase, inviteLink string, chatID int64) (*core.Record, error) {
	prefix, redacted := strings.CutSuffix(inviteLink, "...")
	if redacted {
		return inviteTokenByPrefix(app, prefix, chatID)
	}

	return app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = {:service}",
		map[string]any{
			"token":   inviteLink,
			"service": inviteTokenService,
		},
	)
}

func inviteTokenByPrefix(app *pocketbase.PocketBase, prefix string, chatID int64) (*core.Record, error) {
	if strings.TrimSpace(prefix) == "" || chatID == 0 {
		return nil, nil
	}

	group, err := groupByTelegramChatID(app, chatID)
	if err != nil || group == nil {
		return nil, err
	}

	tokens, err := app.FindRecordsByFilter(
		"tokens",
		"group = {:group} && service = {:service}",
		"",
		0,
		0,
		map[string]any{
			"group":   group.Id,
			"service": inviteTokenService,
		},
	)
	if err != nil {
		return nil, err
	}

	var match *core.Record
	for _, token := range tokens {
		if strings.HasPrefix(strings.TrimSpace(token.GetString("token")), prefix) {
			if match != nil {
				return nil, fmt.Errorf("ambiguous redacted invite link")
			}
			match = token
		}
	}
	return match, nil
}

func groupByTelegramChatID(app *pocketbase.PocketBase, chatID int64) (*core.Record, error) {
	groups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
	if err != nil {
		return nil, err
	}
	needle := strconv.FormatInt(chatID, 10)
	for _, group := range groups {
		telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
		if strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"])) == needle {
			return group, nil
		}
	}
	return nil, nil
}

func GenerateGroupInvite(app *pocketbase.PocketBase, user *core.Record, group *core.Record) (string, error) {
	if user == nil {
		return "", apis.NewBadRequestError("missing_user", nil)
	}
	if strings.TrimSpace(user.GetString("status")) == "cancelled" {
		return "", apis.NewForbiddenError("user_cancelled", nil)
	}
	if group == nil {
		return "", apis.NewBadRequestError("missing_group", nil)
	}

	tg := currentBot()
	if tg == nil {
		return "", apis.NewBadRequestError("telegram_bot_unavailable", fmt.Errorf("telegram bot is not initialized"))
	}

	chatID, err := TelegramChatIDForGroup(group)
	if err != nil {
		return "", err
	}

	userData := backendinternal.ParseJSONMap(user.Get("data"))
	inviteName := strings.TrimSpace(backendinternal.AnyToString(userData["full_name"]))
	if inviteName == "" {
		inviteName = strings.TrimSpace(user.Id)
	}
	if len([]rune(inviteName)) > 32 {
		inviteName = strings.TrimSpace(user.Id)
	}

	invite, err := createChatInviteLink(tg, chatID, inviteName, true, time.Time{})
	if err != nil {
		log.Printf("[group-invite] user=%s group=%s chat_id=%d createChatInviteLink error: %v", strings.TrimSpace(user.Id), strings.TrimSpace(group.Id), chatID, err)
		return "", apis.NewBadRequestError("invite_link_generation_failed", err)
	}

	link := strings.TrimSpace(invite.InviteLink)
	if link == "" {
		return "", apis.NewBadRequestError("invite_link_generation_failed", fmt.Errorf("empty invite link"))
	}
	if invite.IsRevoked {
		return "", apis.NewBadRequestError("invite_link_generation_failed", fmt.Errorf("created invite link is revoked"))
	}
	if err := GenerateInviteToken(app, user.Id, group.Id, link); err != nil {
		_, _ = tg.Request(tgbotapi.RevokeChatInviteLinkConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
			InviteLink: link,
		})
		return "", err
	}

	return link, nil
}

func InviteLinkForUserGroup(app *pocketbase.PocketBase, userID string, groupID string) (string, error) {
	userID = strings.TrimSpace(userID)
	groupID = strings.TrimSpace(groupID)
	if userID == "" || groupID == "" {
		return "", nil
	}

	record, err := app.FindFirstRecordByFilter(
		"tokens",
		"user = {:user} && service = {:service} && group = {:group}",
		map[string]any{
			"user":    userID,
			"service": inviteTokenService,
			"group":   groupID,
		},
	)
	if err != nil || record == nil {
		return "", err
	}
	return strings.TrimSpace(record.GetString("token")), nil
}

func RevokeInviteLink(chatID int64, inviteLink string) error {
	tg := currentBot()
	if tg == nil {
		return nil
	}
	if strings.TrimSpace(inviteLink) == "" {
		return nil
	}

	_, err := tg.Request(tgbotapi.RevokeChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		InviteLink: strings.TrimSpace(inviteLink),
	})
	return err
}

func localGroupInviteTarget(app *pocketbase.PocketBase) (*core.Record, int64, error) {
	group, err := app.FindFirstRecordByFilter("groups", "type = 'local'", map[string]any{})
	if err != nil || group == nil {
		return nil, 0, apis.NewNotFoundError("local_group_not_found", err)
	}

	telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
	chatIDRaw := strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"]))
	if chatIDRaw == "" {
		return nil, 0, apis.NewBadRequestError("invalid_default_group", fmt.Errorf("missing telegram.chat_id"))
	}

	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return nil, 0, apis.NewBadRequestError("invalid_default_group", fmt.Errorf("invalid telegram.chat_id"))
	}
	return group, chatID, nil
}

func TelegramChatIDForGroup(group *core.Record) (int64, error) {
	if group == nil {
		return 0, apis.NewBadRequestError("missing_group", nil)
	}

	telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
	chatIDRaw := strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"]))
	if chatIDRaw == "" {
		return 0, apis.NewBadRequestError("invalid_group", fmt.Errorf("missing telegram.chat_id"))
	}

	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return 0, apis.NewBadRequestError("invalid_group", fmt.Errorf("invalid telegram.chat_id"))
	}
	return chatID, nil
}

func createChatInviteLink(tg *tgbotapi.BotAPI, chatID int64, name string, createsJoinRequest bool, expireAt time.Time) (tgbotapi.ChatInviteLink, error) {
	config := tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig:         tgbotapi.ChatConfig{ChatID: chatID},
		CreatesJoinRequest: createsJoinRequest,
	}
	if strings.TrimSpace(name) != "" {
		config.Name = strings.TrimSpace(name)
	}
	if !expireAt.IsZero() {
		config.ExpireDate = int(expireAt.Unix())
	}

	resp, err := tg.Request(config)
	if err != nil {
		return tgbotapi.ChatInviteLink{}, err
	}

	var invite tgbotapi.ChatInviteLink
	if err := json.Unmarshal(resp.Result, &invite); err != nil {
		return tgbotapi.ChatInviteLink{}, err
	}
	return invite, nil
}

func logMembershipDiagnostics(tg *tgbotapi.BotAPI, actor *core.Record, chatID int64) {
	authTelegram := backendinternal.ParseJSONMap(actor.Get("telegram"))
	authTelegramID, ok := backendinternal.AnyToInt64(authTelegram["id"])
	if ok {
		member, memberErr := tg.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: chatID, UserID: authTelegramID},
		})
		if memberErr != nil {
			log.Printf("[default-invite] user=%s tg_user=%d getChatMember error: %v", strings.TrimSpace(actor.Id), authTelegramID, memberErr)
		} else if member.User != nil {
			log.Printf("[default-invite] user=%s tg_user=%d member_status=%s is_member=%v", strings.TrimSpace(actor.Id), authTelegramID, member.Status, member.Status == "member" || member.Status == "administrator" || member.Status == "creator")
		} else {
			log.Printf("[default-invite] user=%s tg_user=%d member_status=%s", strings.TrimSpace(actor.Id), authTelegramID, member.Status)
		}
		return
	}
	if authTelegram["id"] != nil {
		log.Printf("[default-invite] user=%s telegram.id not parseable: %v", strings.TrimSpace(actor.Id), authTelegram["id"])
	}
}
