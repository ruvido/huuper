package groups

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func FindByPathID(app *pocketbase.PocketBase, e *core.RequestEvent) (*core.Record, error) {
	id := strings.TrimSpace(e.Request.PathValue("id"))
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_group", nil)
	}
	group, err := app.FindRecordById("groups", id)
	if err != nil || group == nil {
		return nil, apis.NewNotFoundError("group_not_found", err)
	}
	return group, nil
}

func UserDisplayName(user *core.Record) string {
	if user == nil {
		return ""
	}
	data := backendinternal.ParseJSONMap(user.Get("data"))
	if fullName, ok := data["full_name"].(string); ok && strings.TrimSpace(fullName) != "" {
		return strings.TrimSpace(fullName)
	}
	return strings.TrimSpace(user.GetString("email"))
}

func usersByID(app *pocketbase.PocketBase, userIDs []string) (map[string]*core.Record, error) {
	if len(userIDs) == 0 {
		return map[string]*core.Record{}, nil
	}

	seen := make(map[string]struct{}, len(userIDs))
	deduped := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		deduped = append(deduped, trimmed)
	}

	records, err := app.FindRecordsByIds("users", deduped)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*core.Record, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		out[record.Id] = record
	}

	return out, nil
}

func groupsByID(app *pocketbase.PocketBase, groupIDs []string) (map[string]*core.Record, error) {
	if len(groupIDs) == 0 {
		return map[string]*core.Record{}, nil
	}

	seen := make(map[string]struct{}, len(groupIDs))
	deduped := make([]string, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		trimmed := strings.TrimSpace(groupID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		deduped = append(deduped, trimmed)
	}

	records, err := app.FindRecordsByIds("groups", deduped)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*core.Record, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		out[record.Id] = record
	}

	return out, nil
}
