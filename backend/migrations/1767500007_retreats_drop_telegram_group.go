package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Drops retreats.telegram_group. A retreat's Telegram chat is a one-off group
// that lives for a few days around the gathering: the bot is not in it, so
// there is no `groups` record to relate to and no chat to mint per-member
// invite links from. The organiser pastes one invite link into
// `data.telegram_invite` instead — see retreats.InviteLinkForRetreat.
//
// The relation was never populated on any retreat, so nothing is lost here.
func init() {
	m.Register(func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return err
		}

		field := retreats.Fields.GetByName("telegram_group")
		if field == nil {
			return nil
		}
		retreats.Fields.RemoveById(field.GetId())

		return app.Save(retreats)
	}, func(app core.App) error {
		retreats, err := app.FindCollectionByNameOrId("retreats")
		if err != nil {
			return err
		}
		if retreats.Fields.GetByName("telegram_group") != nil {
			return nil
		}

		groups, err := app.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}

		retreats.Fields.Add(&core.RelationField{
			Name:         "telegram_group",
			Required:     false,
			CollectionId: groups.Id,
			MaxSelect:    1,
		})

		return app.Save(retreats)
	})
}
