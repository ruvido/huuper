package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"members/bot"
	backendinternal "members/internal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func DefaultGroupInvite(app *pocketbase.PocketBase, actor *core.Record) (string, string, error) {
	if actor == nil {
		return "", "", apis.NewUnauthorizedError("Unauthorized", nil)
	}
	log.Printf("[default-invite] user=%s request started", strings.TrimSpace(actor.Id))

	group, err := app.FindFirstRecordByFilter("groups", "type = 'default'", map[string]any{})
	if err != nil || group == nil {
		return "", "", apis.NewNotFoundError("default_group_not_found", err)
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

	tg := bot.GetBot()
	if tg == nil {
		return "", "", apis.NewBadRequestError("telegram_bot_unavailable", fmt.Errorf("telegram bot is not initialized"))
	}

	logMembershipDiagnostics(tg, actor, chatID)

	resp, err := tg.Request(tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig:         tgbotapi.ChatConfig{ChatID: chatID},
		CreatesJoinRequest: false,
	})
	if err != nil {
		log.Printf("[default-invite] user=%s chat_id=%d createChatInviteLink error: %v", strings.TrimSpace(actor.Id), chatID, err)
		return "", "", apis.NewBadRequestError("invite_link_generation_failed", err)
	}

	var invite tgbotapi.ChatInviteLink
	if err := json.Unmarshal(resp.Result, &invite); err != nil {
		log.Printf("[default-invite] user=%s chat_id=%d createChatInviteLink unmarshal error: %v", strings.TrimSpace(actor.Id), chatID, err)
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
