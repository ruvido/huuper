package retreats

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InviteLinkForRetreat returns the retreat's Telegram invite link, taken from
// `data.telegram_invite` on the record. A retreat group is a one-off chat that
// exists for a few days around the event: unlike the standing Realmen groups
// it is not run by the bot, so there is no chat to mint per-member links from
// — the organiser pastes the invite link and everyone gets the same one.
// Returns "" when unset, which only leaves the confirmation email without its
// button; it must never block activation.
func InviteLinkForRetreat(app *pocketbase.PocketBase, retreat *core.Record) string {
	if retreat == nil {
		return ""
	}
	data := parseData(retreat)
	link, _ := data["telegram_invite"].(string)
	link = strings.TrimSpace(link)
	if link == "" {
		app.Logger().Warn("retreats: no telegram invite link set", "retreat", retreat.Id)
	}
	return link
}
