package requests

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func buildPromotedUserData(app *pocketbase.PocketBase, record *core.Record, data map[string]any) map[string]any {
	userData := BuildUserData(data)
	userData["request"] = promotedRequestSnapshot(app, record, data)
	if rawSnapshot, ok := data[RequestFlowDataKey].(map[string]any); ok {
		userData["request_flow"] = backendinternal.DeepCopyJSONMap(rawSnapshot)
	}
	return userData
}

func promotedRequestSnapshot(app *pocketbase.PocketBase, record *core.Record, data map[string]any) map[string]any {
	if record == nil {
		return nil
	}

	snapshot := map[string]any{
		"id": record.Id,
	}

	if group := groupAssignmentSnapshot(app, record, data); group != nil {
		snapshot["group_assignment"] = group
	}
	if guardian := guardianAssignmentSnapshot(app, record, data); guardian != nil {
		snapshot["guardian_assignment"] = guardian
	}
	if mentoring := mentoringSnapshot(data); mentoring != nil {
		snapshot["mentoring"] = mentoring
	}
	if approval := approvalSnapshot(data, "group_approved_at", "group_approved_by"); approval != nil {
		snapshot["group_approval"] = approval
	}
	if approval := approvalSnapshot(data, "admin_approved_at", "admin_approved_by"); approval != nil {
		snapshot["admin_approval"] = approval
	}

	return snapshot
}

func groupAssignmentSnapshot(app *pocketbase.PocketBase, record *core.Record, data map[string]any) map[string]any {
	groupID := strings.TrimSpace(record.GetString("group"))
	if groupID == "" {
		return nil
	}

	block := mapFromData(data, "assign_group")
	out := eventSnapshot(block, "assigned_at", "assigned_by")
	name := groupName(app, groupID)
	if name == "" {
		name = relationNameFromBlock(block, "group")
	}
	out["group"] = relationSnapshot(groupID, name)
	return out
}

func guardianAssignmentSnapshot(app *pocketbase.PocketBase, record *core.Record, data map[string]any) map[string]any {
	guardianID := strings.TrimSpace(record.GetString("guardian"))
	if guardianID == "" {
		return nil
	}

	block := mapFromData(data, "guardian")
	out := eventSnapshot(block, "assigned_at", "assigned_by")
	name := strings.TrimSpace(backendinternal.AnyToString(block["name"]))
	if name == "" {
		name = userName(app, guardianID)
	}
	out["guardian"] = relationSnapshot(guardianID, name)
	return out
}

func mentoringSnapshot(data map[string]any) map[string]any {
	block := mapFromData(data, mentoringDataKey)
	out := eventSnapshot(block, "done_at", "done_by")
	if len(out) == 0 {
		return nil
	}
	if notes, ok := block["notes"]; ok {
		out["notes"] = notes
	}
	return out
}

func approvalSnapshot(data map[string]any, atKey string, byKey string) map[string]any {
	out := eventSnapshot(data, atKey, byKey)
	if len(out) == 0 {
		return nil
	}
	return out
}

func eventSnapshot(data map[string]any, atKey string, byKey string) map[string]any {
	out := map[string]any{}
	if at := strings.TrimSpace(backendinternal.AnyToString(data[atKey])); at != "" {
		out["at"] = at
	}
	if by := strings.TrimSpace(backendinternal.AnyToString(data[byKey])); by != "" {
		out["by"] = by
	}
	return out
}

func relationSnapshot(id string, name string) map[string]any {
	return map[string]any{
		"id":   strings.TrimSpace(id),
		"name": strings.TrimSpace(name),
	}
}

func relationNameFromBlock(data map[string]any, key string) string {
	block, _ := data[key].(map[string]any)
	if block == nil {
		return ""
	}
	return strings.TrimSpace(backendinternal.AnyToString(block["name"]))
}

func mapFromData(data map[string]any, key string) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	block, _ := data[key].(map[string]any)
	if block == nil {
		return map[string]any{}
	}
	return block
}

func groupName(app *pocketbase.PocketBase, groupID string) string {
	if app == nil || strings.TrimSpace(groupID) == "" {
		return ""
	}
	group, err := app.FindRecordById("groups", strings.TrimSpace(groupID))
	if err != nil || group == nil {
		return ""
	}
	return strings.TrimSpace(group.GetString("name"))
}

func userName(app *pocketbase.PocketBase, userID string) string {
	if app == nil || strings.TrimSpace(userID) == "" {
		return ""
	}
	user, err := app.FindRecordById("users", strings.TrimSpace(userID))
	if err != nil || user == nil {
		return ""
	}
	return actorDisplayName(user)
}
