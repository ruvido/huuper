package groups

import (
	"slices"
	"strings"

	tginternal "members/backend/internal/telegram"

	"github.com/pocketbase/pocketbase"
)

func ListForUser(app *pocketbase.PocketBase, userID string) ([]GroupListItem, error) {
	userID = strings.TrimSpace(userID)

	membershipRelations, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"created",
		500,
		0,
		map[string]any{"user": userID},
	)
	if err != nil {
		return nil, err
	}

	tokenRelations, err := app.FindRecordsByFilter(
		"tokens",
		"user = {:user} && service = {:service}",
		"created",
		500,
		0,
		map[string]any{
			"user":    userID,
			"service": "telegram_invite",
		},
	)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, 0, len(membershipRelations)+len(tokenRelations))
	groupTokens := make(map[string]string, len(tokenRelations))
	memberships := make(map[string]struct{}, len(membershipRelations))
	seen := make(map[string]struct{}, len(membershipRelations)+len(tokenRelations))

	appendGroupID := func(groupID string) {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" {
			return
		}
		if _, ok := seen[groupID]; ok {
			return
		}
		seen[groupID] = struct{}{}
		groupIDs = append(groupIDs, groupID)
	}

	for _, rel := range membershipRelations {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID == "" {
			continue
		}
		memberships[groupID] = struct{}{}
		appendGroupID(groupID)
	}
	for _, rel := range tokenRelations {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID == "" {
			continue
		}
		appendGroupID(groupID)
		if token := strings.TrimSpace(rel.GetString("token")); token != "" {
			groupTokens[groupID] = token
		}
	}

	groupsByID, err := groupsByID(app, groupIDs)
	if err != nil {
		return nil, err
	}

	filteredGroupIDs := make([]string, 0, len(groupIDs))
	filteredTokens := make(map[string]string, len(groupTokens))
	for _, groupID := range groupIDs {
		group := groupsByID[groupID]
		if group == nil {
			continue
		}
		if _, err := tginternal.TelegramChatIDForGroup(group); err != nil {
			continue
		}
		filteredGroupIDs = append(filteredGroupIDs, groupID)
		if token := strings.TrimSpace(groupTokens[groupID]); token != "" {
			filteredTokens[groupID] = token
		}
	}

	memberCounts, err := memberCountsByGroup(app, filteredGroupIDs)
	if err != nil {
		return nil, err
	}

	requestCounts, err := requestCountsByGroup(app, filteredGroupIDs)
	if err != nil {
		return nil, err
	}

	items := make([]GroupListItem, 0, len(filteredGroupIDs))
	for _, groupID := range filteredGroupIDs {
		group := groupsByID[groupID]
		if group == nil {
			continue
		}

		var visibleRequestsCount *int
		if strings.TrimSpace(group.GetString("assistant")) == strings.TrimSpace(userID) {
			count := requestCounts[groupID]
			visibleRequestsCount = &count
		}

		_, isMember := memberships[groupID]
		items = append(items, GroupListItem{
			ID:            group.Id,
			Name:          strings.TrimSpace(group.GetString("name")),
			Type:          strings.TrimSpace(group.GetString("type")),
			Description:   strings.TrimSpace(group.GetString("description")),
			Assistant:     strings.TrimSpace(group.GetString("assistant")),
			MembersCount:  memberCounts[groupID],
			RequestsCount: visibleRequestsCount,
			InviteLink:    filteredTokens[groupID],
			IsMember:      isMember,
		})
	}

	slices.SortStableFunc(items, func(a GroupListItem, b GroupListItem) int {
		if cmp := compareGroupType(a.Type, b.Type); cmp != 0 {
			return cmp
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return items, nil
}

func memberCountsByGroup(app *pocketbase.PocketBase, groupIDs []string) (map[string]int, error) {
	out := make(map[string]int, len(groupIDs))
	for _, groupID := range groupIDs {
		records, err := app.FindRecordsByFilter(
			"user_groups",
			"group = {:group}",
			"",
			500,
			0,
			map[string]any{"group": groupID},
		)
		if err != nil {
			return nil, err
		}
		out[groupID] = len(records)
	}
	return out, nil
}

func requestCountsByGroup(app *pocketbase.PocketBase, groupIDs []string) (map[string]int, error) {
	out := make(map[string]int, len(groupIDs))
	for _, groupID := range groupIDs {
		records, err := app.FindRecordsByFilter(
			"requests",
			"group = {:group} && archived = false",
			"",
			500,
			0,
			map[string]any{"group": groupID},
		)
		if err != nil {
			return nil, err
		}
		out[groupID] = len(records)
	}
	return out, nil
}

func compareGroupType(a string, b string) int {
	return groupTypeRank(strings.TrimSpace(a)) - groupTypeRank(strings.TrimSpace(b))
}

func groupTypeRank(value string) int {
	switch value {
	case "default":
		return 0
	case "local":
		return 1
	default:
		return 2
	}
}
