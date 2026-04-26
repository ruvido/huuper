package groups

import (
	"sort"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

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
	group, err := app.FindRecordById("groups", groupID)
	if err != nil || group == nil {
		return nil, err
	}
	assistantID := strings.TrimSpace(group.GetString("assistant"))

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
	for _, rel := range relations {
		userID := strings.TrimSpace(rel.GetString("user"))
		if userID == "" {
			continue
		}
		userIDs = append(userIDs, userID)
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
		isAssistant := assistantID != "" && user.Id == assistantID
		_, isGuardian := guardianCounts[userID]
		if !shouldIncludeGroupMember(user, isAssistant, isGuardian) {
			continue
		}
		items = append(items, MemberItem{
			ID:            user.Id,
			Email:         user.GetString("email"),
			FullName:      UserDisplayName(user),
			Avatar:        strings.TrimSpace(user.GetString("avatar")),
			Age:           UserAge(user),
			Region:        UserRegion(user),
			IsAssistant:   isAssistant,
			IsGuardian:    isGuardian,
			ProtegesCount: guardianCounts[userID],
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsAssistant != items[j].IsAssistant {
			return items[i].IsAssistant
		}
		if items[i].IsGuardian != items[j].IsGuardian {
			return items[i].IsGuardian
		}

		left := strings.ToLower(strings.TrimSpace(items[i].FullName))
		right := strings.ToLower(strings.TrimSpace(items[j].FullName))
		if left == "" {
			left = strings.ToLower(strings.TrimSpace(items[i].Email))
		}
		if right == "" {
			right = strings.ToLower(strings.TrimSpace(items[j].Email))
		}
		return left < right
	})

	return &MembersResponse{
		GroupID: groupID,
		Items:   items,
	}, nil
}

func shouldIncludeGroupMember(user *core.Record, isAssistant bool, isGuardian bool) bool {
	if user == nil {
		return false
	}
	if !isAdminUser(user) {
		return true
	}
	return isAssistant || isGuardian
}

func isAdminUser(user *core.Record) bool {
	if user == nil {
		return false
	}
	return user.GetBool("admin")
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
			Avatar:        strings.TrimSpace(user.GetString("avatar")),
			ProtegesCount: guardianCounts[userID],
		})
	}

	return &GuardiansResponse{
		GroupID: groupID,
		Items:   items,
	}, nil
}

func GuardianGroupsForUser(app *pocketbase.PocketBase, userID string) ([]GuardianGroupItem, error) {
	records, err := app.FindRecordsByFilter(
		"requests",
		"guardian = {:guardian} && rejected = false",
		"",
		500,
		0,
		map[string]any{"guardian": strings.TrimSpace(userID)},
	)
	if err != nil {
		return nil, err
	}

	countsByGroup := map[string]int{}
	groupIDs := make([]string, 0)
	for _, record := range records {
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			continue
		}
		if _, ok := countsByGroup[groupID]; !ok {
			groupIDs = append(groupIDs, groupID)
		}
		countsByGroup[groupID]++
	}

	groups, err := groupsByID(app, groupIDs)
	if err != nil {
		return nil, err
	}

	items := make([]GuardianGroupItem, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group := groups[groupID]
		if group == nil {
			continue
		}
		items = append(items, GuardianGroupItem{
			ID:            group.Id,
			Name:          strings.TrimSpace(group.GetString("name")),
			ProtegesCount: countsByGroup[groupID],
		})
	}

	return items, nil
}
