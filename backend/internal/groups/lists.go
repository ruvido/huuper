package groups

import (
	"slices"
	"strings"

	"github.com/pocketbase/pocketbase"
)

func ListForUser(app *pocketbase.PocketBase, userID string) ([]GroupListItem, error) {
	relations, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"created",
		500,
		0,
		map[string]any{"user": strings.TrimSpace(userID)},
	)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, 0, len(relations))
	for _, rel := range relations {
		groupID := strings.TrimSpace(rel.GetString("group"))
		if groupID == "" {
			continue
		}
		groupIDs = append(groupIDs, groupID)
	}

	groupsByID, err := groupsByID(app, groupIDs)
	if err != nil {
		return nil, err
	}

	memberCounts, err := memberCountsByGroup(app, groupIDs)
	if err != nil {
		return nil, err
	}

	requestCounts, err := requestCountsByGroup(app, groupIDs)
	if err != nil {
		return nil, err
	}

	items := make([]GroupListItem, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group := groupsByID[groupID]
		if group == nil {
			continue
		}

		var visibleRequestsCount *int
		if strings.TrimSpace(group.GetString("assistant")) == strings.TrimSpace(userID) {
			count := requestCounts[groupID]
			visibleRequestsCount = &count
		}

		items = append(items, GroupListItem{
			ID:            group.Id,
			Name:          strings.TrimSpace(group.GetString("name")),
			Type:          strings.TrimSpace(group.GetString("type")),
			Description:   strings.TrimSpace(group.GetString("description")),
			Assistant:     strings.TrimSpace(group.GetString("assistant")),
			MembersCount:  memberCounts[groupID],
			RequestsCount: visibleRequestsCount,
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
			"group = {:group} && rejected = false",
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
