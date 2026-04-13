package groups

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
)

func UpdateAssistant(app *pocketbase.PocketBase, groupID string, assistantID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return apis.NewBadRequestError("invalid_group", nil)
	}

	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return apis.NewNotFoundError("group_not_found", err)
	}

	assistantID = strings.TrimSpace(assistantID)
	if assistantID == "" {
		return apis.NewBadRequestError("invalid_assistant", nil)
	}

	assistant, err := app.FindRecordById("users", assistantID)
	if err != nil || assistant == nil {
		return apis.NewNotFoundError("assistant_not_found", err)
	}

	ok, err := backendinternal.IsMemberOfGroup(app, assistantID, groupID)
	if err != nil {
		return apis.NewBadRequestError("failed_group_membership_check", err)
	}
	if !ok {
		return apis.NewBadRequestError("assistant_must_be_group_member", nil)
	}

	group.Set("assistant", assistantID)
	if err := app.Save(group); err != nil {
		return apis.NewBadRequestError("failed_update_group", err)
	}

	return nil
}
