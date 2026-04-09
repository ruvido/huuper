package requests

import (
	"strings"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
)

func GuardianRequestsForUser(app *pocketbase.PocketBase, userID string, visibleGroupIDs map[string]struct{}) ([]GuardianRequestItem, error) {
	records, err := app.FindRecordsByFilter(
		"requests",
		"guardian = {:guardian} && rejected = false",
		"-updated",
		500,
		0,
		map[string]any{"guardian": strings.TrimSpace(userID)},
	)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, 0, len(records))
	seenGroups := map[string]struct{}{}
	for _, record := range records {
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			continue
		}
		if visibleGroupIDs != nil {
			if _, ok := visibleGroupIDs[groupID]; !ok {
				continue
			}
		}
		if _, ok := seenGroups[groupID]; ok {
			continue
		}
		seenGroups[groupID] = struct{}{}
		groupIDs = append(groupIDs, groupID)
	}

	groupsByID := map[string]string{}
	if len(groupIDs) > 0 {
		groups, err := app.FindRecordsByIds("groups", groupIDs)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			if group == nil {
				continue
			}
			groupsByID[group.Id] = strings.TrimSpace(group.GetString("name"))
		}
	}

	items := make([]GuardianRequestItem, 0, len(records))
	for _, record := range records {
		groupID := strings.TrimSpace(record.GetString("group"))
		if groupID == "" {
			continue
		}
		if visibleGroupIDs != nil {
			if _, ok := visibleGroupIDs[groupID]; !ok {
				continue
			}
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		flow, err := LoadFlowForRequest(app, data)
		if err != nil {
			return nil, err
		}
		stepIndex := EffectiveStepIndex(record, data, flow)
		status := StatusForItem(record.GetBool("rejected"), stepIndex, flow.Steps)
		statusLabel := status
		if nextStep, hasNext := FlowStepAt(flow, stepIndex); hasNext && strings.TrimSpace(nextStep.Label) != "" {
			statusLabel = strings.TrimSpace(nextStep.Label)
		}
		guardianData, _ := data["guardian"].(map[string]any)
		assignedAt, _ := guardianData["assigned_at"].(string)

		items = append(items, GuardianRequestItem{
			ID:          record.Id,
			FullName:    strings.TrimSpace(DisplayName(data, strings.TrimSpace(record.GetString("email")), record.Id)),
			Email:       strings.TrimSpace(record.GetString("email")),
			Status:      strings.TrimSpace(status),
			StatusLabel: strings.TrimSpace(statusLabel),
			GroupID:     groupID,
			GroupName:   groupsByID[groupID],
			Created:     strings.TrimSpace(record.GetString("created")),
			AssignedAt:  strings.TrimSpace(assignedAt),
		})
	}

	return items, nil
}
