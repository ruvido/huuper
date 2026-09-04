package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

// CreateInviteLinkForGroup generates a fresh Telegram invite link for the
// given `groups` record, reusing the same createChatInviteLink primitive
// DefaultGroupInvite uses. Added as a standalone call-site for the retreats
// module (a retreat's telegram_group points at a `groups` record populated
// exactly the way any other group is: bot promoted to admin in a Telegram
// group creates the `groups` record with its chat_id automatically).
func CreateInviteLinkForGroup(app *pocketbase.PocketBase, groupID string) (string, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return "", fmt.Errorf("missing group")
	}

	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return "", fmt.Errorf("group not found: %w", err)
	}

	telegramData := backendinternal.ParseJSONMap(group.Get("telegram"))
	chatIDRaw := strings.TrimSpace(backendinternal.AnyToString(telegramData["chat_id"]))
	if chatIDRaw == "" {
		return "", fmt.Errorf("group has no telegram.chat_id")
	}

	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid telegram.chat_id")
	}

	tg := currentBot()
	if tg == nil {
		return "", fmt.Errorf("telegram bot is not initialized")
	}

	invite, err := createChatInviteLink(tg, chatID, "", false, time.Time{})
	if err != nil {
		return "", err
	}

	link := strings.TrimSpace(invite.InviteLink)
	if link == "" {
		return "", fmt.Errorf("empty invite link")
	}
	if invite.IsRevoked {
		return "", fmt.Errorf("created invite link is revoked")
	}

	return link, nil
}
