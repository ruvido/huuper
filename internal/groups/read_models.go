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

type MembersResponse struct {
	GroupID string       `json:"group_id"`
	Items   []MemberItem `json:"items"`
}

type GuardiansResponse struct {
	GroupID string         `json:"group_id"`
	Items   []GuardianItem `json:"items"`
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

func MembersResponseForGroup(app *pocketbase.PocketBase, groupID string) (*MembersResponse, error) {
	relations, err := app.FindRecordsByFilter(
		"user_groups",
		"group = {:group}",
		"created",
		500,
		0,
		map[string]any{"group": groupID},
	)
	if err != nil {
		return nil, err
	}

	guardianCounts, err := GuardianCounts(app, groupID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(relations))
	rolesByUserID := make(map[string]string, len(relations))
	for _, rel := range relations {
		userID := strings.TrimSpace(rel.GetString("user"))
		if userID == "" {
			continue
		}
		userIDs = append(userIDs, userID)
		rolesByUserID[userID] = strings.TrimSpace(rel.GetString("role"))
	}

	usersByID, err := usersByID(app, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]MemberItem, 0, len(relations))
	for _, userID := range userIDs {
		user := usersByID[userID]
		if user == nil {
			continue
		}
		_, isGuardian := guardianCounts[userID]
		items = append(items, MemberItem{
			ID:         user.Id,
			Email:      user.GetString("email"),
			FullName:   UserDisplayName(user),
			Role:       rolesByUserID[userID],
			IsGuardian: isGuardian,
		})
	}

	return &MembersResponse{
		GroupID: groupID,
		Items:   items,
	}, nil
}

func GuardiansResponseForGroup(app *pocketbase.PocketBase, groupID string) (*GuardiansResponse, error) {
	guardianCounts, err := GuardianCounts(app, groupID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(guardianCounts))
	for userID := range guardianCounts {
		userIDs = append(userIDs, userID)
	}

	usersByID, err := usersByID(app, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]GuardianItem, 0, len(guardianCounts))
	for _, userID := range userIDs {
		user := usersByID[userID]
		if user == nil {
			continue
		}
		items = append(items, GuardianItem{
			ID:            user.Id,
			Email:         user.GetString("email"),
			FullName:      UserDisplayName(user),
			ProtegesCount: guardianCounts[userID],
		})
	}

	return &GuardiansResponse{
		GroupID: groupID,
		Items:   items,
	}, nil
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
