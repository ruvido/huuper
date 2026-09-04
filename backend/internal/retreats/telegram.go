package retreats

import (
	"strings"

	telegraminternal "members/backend/internal/telegram"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InviteLinkForRetreat returns a fresh Telegram invite link for the
// retreat's linked group, or "" if the retreat has no telegram_group set
// (or link generation failed — this must never block registration
// activation, it only enriches the confirmation email).
func InviteLinkForRetreat(app *pocketbase.PocketBase, retreat *core.Record) string {
	if retreat == nil {
		return ""
	}
	groupID := strings.TrimSpace(retreat.GetString("telegram_group"))
	if groupID == "" {
		return ""
	}
	link, err := telegraminternal.CreateInviteLinkForGroup(app, groupID)
	if err != nil {
		app.Logger().Warn("retreats: failed to create telegram invite link", "error", err, "group", groupID)
		return ""
	}
	return link
}
