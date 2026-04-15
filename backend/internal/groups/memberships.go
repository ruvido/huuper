package groups

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func EnsureMembership(app *pocketbase.PocketBase, userID string, groupID string) error {
	userID = strings.TrimSpace(userID)
	groupID = strings.TrimSpace(groupID)
	if userID == "" || groupID == "" {
		return nil
	}

	existing, err := app.FindFirstRecordByFilter(
		"user_groups",
		"user = {:user} && group = {:group}",
		map[string]any{
			"user":  userID,
			"group": groupID,
		},
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if existing != nil {
		return nil
	}

	collection, err := app.FindCollectionByNameOrId("user_groups")
	if err != nil {
		return err
	}

	record := core.NewRecord(collection)
	record.Set("user", userID)
	record.Set("group", groupID)
	return app.Save(record)
}

func RemoveMembership(app *pocketbase.PocketBase, userID string, groupID string) error {
	userID = strings.TrimSpace(userID)
	groupID = strings.TrimSpace(groupID)
	if userID == "" || groupID == "" {
		return nil
	}

	existing, err := app.FindFirstRecordByFilter(
		"user_groups",
		"user = {:user} && group = {:group}",
		map[string]any{
			"user":  userID,
			"group": groupID,
		},
	)
	if err != nil || existing == nil {
		return err
	}
	return app.Delete(existing)
}
