package groups

import (
	"strings"

	backendinternal "members/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type MemberItem struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"`
	IsGuardian bool   `json:"is_guardian"`
}

type GuardianItem struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	ProtegesCount int    `json:"proteges_count"`
}

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

func GuardianCounts(app *pocketbase.PocketBase, groupID string) (map[string]int, error) {
	records, err := app.FindRecordsByFilter(
		"requests",
		"group = {:group} && guardian != '' && rejected = false",
		"",
		500,
		0,
		map[string]any{"group": groupID},
	)
	if err != nil {
		return nil, err
	}

	out := map[string]int{}
	for _, record := range records {
		guardianID := strings.TrimSpace(record.GetString("guardian"))
		if guardianID == "" {
			continue
		}
		out[guardianID]++
	}
	return out, nil
}
