package events

import (
	"sort"
	"strconv"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	groupinternal "members/backend/internal/groups"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type AttendeeItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id,omitempty"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Avatar    string `json:"avatar,omitempty"`
	Age       *int   `json:"age,omitempty"`
	AgeRange  string `json:"age_range,omitempty"`
	Region    string `json:"region,omitempty"`
	GroupName string `json:"group_name,omitempty"`
	Status    string `json:"status,omitempty"`
	IsUser    bool   `json:"is_user"`
	IsAdmin   bool   `json:"is_admin"`
}

func ActiveAttendeesForEvent(app *pocketbase.PocketBase, eventID string) ([]AttendeeItem, error) {
	return attendeesForEvent(app, eventID, []string{"active"})
}

func PendingAttendeesForEvent(app *pocketbase.PocketBase, eventID string) ([]AttendeeItem, error) {
	return attendeesForEvent(app, eventID, []string{"pending"})
}

func CancelledAttendeesForEvent(app *pocketbase.PocketBase, eventID string) ([]AttendeeItem, error) {
	return attendeesForEvent(app, eventID, []string{"rejected", "cancelled"})
}

func attendeesForEvent(app *pocketbase.PocketBase, eventID string, statuses []string) ([]AttendeeItem, error) {
	filterParts := []string{"event = {:event}"}
	params := map[string]any{"event": strings.TrimSpace(eventID)}
	for index, status := range statuses {
		key := "status_" + strconv.Itoa(index)
		filterParts = append(filterParts, "status = {:"+key+"}")
		params[key] = strings.TrimSpace(status)
	}
	filter := strings.Join(filterParts[:1], "")
	if len(statuses) > 0 {
		filter += " && (" + strings.Join(filterParts[1:], " || ") + ")"
	}

	records, err := app.FindRecordsByFilter(
		"event_registrations",
		filter,
		"created",
		0,
		0,
		params,
	)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(records))
	for _, record := range records {
		userID := strings.TrimSpace(record.GetString("user"))
		if userID != "" {
			userIDs = append(userIDs, userID)
		}
	}

	users, err := usersByID(app, userIDs)
	if err != nil {
		return nil, err
	}

	groupNamesByUserID, err := primaryGroupNamesByUserID(app, userIDs)
	if err != nil {
		return nil, err
	}

	items := make([]AttendeeItem, 0, len(records))
	for _, record := range records {
		userID := strings.TrimSpace(record.GetString("user"))
		email := strings.TrimSpace(record.GetString("email"))
		if user, ok := users[userID]; ok && user != nil {
			items = append(items, AttendeeItem{
				ID:        record.Id,
				UserID:    user.Id,
				Email:     strings.TrimSpace(user.GetString("email")),
				FullName:  groupinternal.UserDisplayName(user),
				Avatar:    strings.TrimSpace(user.GetString("avatar")),
				Age:       groupinternal.UserAge(user),
				AgeRange:  userAgeRange(user),
				Region:    groupinternal.UserRegion(user),
				GroupName: groupNamesByUserID[user.Id],
				Status:    strings.TrimSpace(record.GetString("status")),
				IsUser:    true,
				IsAdmin:   user.GetBool("admin"),
			})
			continue
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		items = append(items, AttendeeItem{
			ID:       record.Id,
			Email:    email,
			FullName: registrationDisplayName(data, email, record.Id),
			Age:      registrationAge(data),
			AgeRange: registrationAgeRange(data),
			Region:   registrationRegion(data),
			Status:   strings.TrimSpace(record.GetString("status")),
			IsUser:   false,
		})
	}

	return items, nil
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
		if record != nil {
			out[record.Id] = record
		}
	}
	return out, nil
}

func primaryGroupNamesByUserID(app *pocketbase.PocketBase, userIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(userIDs) == 0 {
		return out, nil
	}

	seen := make(map[string]struct{}, len(userIDs))
	groupIDsByUser := map[string][]string{}
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}

		rels, err := app.FindRecordsByFilter(
			"user_groups",
			"user = {:user}",
			"",
			0,
			0,
			map[string]any{"user": trimmed},
		)
		if err != nil {
			return nil, err
		}

		groupIDs := make([]string, 0, len(rels))
		for _, rel := range rels {
			groupID := strings.TrimSpace(rel.GetString("group"))
			if groupID != "" {
				groupIDs = append(groupIDs, groupID)
			}
		}
		groupIDsByUser[trimmed] = groupIDs
	}

	allGroupIDs := make([]string, 0)
	for _, ids := range groupIDsByUser {
		allGroupIDs = append(allGroupIDs, ids...)
	}
	groups, err := groupsByID(app, allGroupIDs)
	if err != nil {
		return nil, err
	}

	for userID, ids := range groupIDsByUser {
		type candidate struct {
			name     string
			priority int
		}
		candidates := make([]candidate, 0, len(ids))
		for _, groupID := range ids {
			group := groups[groupID]
			if group == nil {
				continue
			}
			name := strings.TrimSpace(group.GetString("name"))
			if name == "" {
				continue
			}
			priority := 2
			switch strings.TrimSpace(group.GetString("type")) {
			case "local":
				priority = 0
			case "default":
				priority = 1
			}
			candidates = append(candidates, candidate{name: name, priority: priority})
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].priority != candidates[j].priority {
				return candidates[i].priority < candidates[j].priority
			}
			return candidates[i].name < candidates[j].name
		})
		if len(candidates) > 0 {
			out[userID] = candidates[0].name
		}
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
		if record != nil {
			out[record.Id] = record
		}
	}
	return out, nil
}

func registrationDisplayName(data map[string]any, email string, fallbackID string) string {
	if fullName, ok := data["full_name"].(string); ok && strings.TrimSpace(fullName) != "" {
		return strings.TrimSpace(fullName)
	}
	if email != "" {
		return email
	}
	return fallbackID
}

func registrationRegion(data map[string]any) string {
	if data == nil {
		return ""
	}
	if region, ok := data["region"].(string); ok && strings.TrimSpace(region) != "" {
		return strings.TrimSpace(region)
	}
	return ""
}

func registrationAgeRange(data map[string]any) string {
	if data == nil {
		return ""
	}
	if ageRange, ok := data["age_range"].(string); ok && strings.TrimSpace(ageRange) != "" {
		return strings.TrimSpace(ageRange)
	}
	return ""
}

func userAgeRange(user *core.Record) string {
	if user == nil {
		return ""
	}
	data := backendinternal.ParseJSONMap(user.Get("data"))
	return registrationAgeRange(data)
}

func registrationAge(data map[string]any) *int {
	if data == nil {
		return nil
	}
	raw, ok := data["birth_year"]
	if !ok {
		return nil
	}

	var birthYear int
	switch value := raw.(type) {
	case float64:
		birthYear = int(value)
	case int:
		birthYear = value
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return nil
		}
		birthYear = parsed
	default:
		return nil
	}

	currentYear := time.Now().Year()
	if birthYear <= 1900 || birthYear > currentYear {
		return nil
	}

	age := currentYear - birthYear
	if age < 0 || age > 120 {
		return nil
	}
	return &age
}
