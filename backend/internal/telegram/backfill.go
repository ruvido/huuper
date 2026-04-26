package telegram

import (
	"log"
	"math"
	"sort"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	copyadmin "members/backend/internal/copywriting/admin"

	"github.com/pocketbase/pocketbase"
)

const sundayEveningStartHour = 20

func BackfillDefaultGroupInvites(app *pocketbase.PocketBase) {
	defaultGroup, err := app.FindFirstRecordByFilter(
		"groups",
		"type = 'default'",
		map[string]any{},
	)
	if err != nil || defaultGroup == nil {
		log.Printf("[backfill] no default group found, skipping: %v", err)
		return
	}

	if _, err := TelegramChatIDForGroup(defaultGroup); err != nil {
		log.Printf("[backfill] default group %s has no telegram chat_id, skipping: %v", defaultGroup.Id, err)
		return
	}

	users, err := app.FindRecordsByFilter(
		"users",
		"status = 'approved'",
		"",
		0,
		0,
		map[string]any{},
	)
	if err != nil {
		log.Printf("[backfill] failed to load approved users: %v", err)
		return
	}

	now := time.Now()
	var total, skippedMember, alreadyToken, created, failed int
	pending := make([]copyadmin.PendingDefaultGroupUser, 0)

	for _, user := range users {
		total++

		memberships, err := app.FindRecordsByFilter(
			"user_groups",
			"user = {:user} && group = {:group}",
			"",
			1,
			0,
			map[string]any{"user": user.Id, "group": defaultGroup.Id},
		)
		if err == nil && len(memberships) > 0 {
			skippedMember++
			continue
		}

		userData := backendinternal.ParseJSONMap(user.Get("data"))
		entry := copyadmin.PendingDefaultGroupUser{
			FullName:   strings.TrimSpace(backendinternal.AnyToString(userData["full_name"])),
			LocalGroup: resolveLocalGroupName(app, user.Id),
			DaysSince:  daysSince(user.GetDateTime("created").Time(), now),
		}

		tokens, err := app.FindRecordsByFilter(
			"tokens",
			"user = {:user} && group = {:group} && service = {:service}",
			"",
			1,
			0,
			map[string]any{
				"user":    user.Id,
				"group":   defaultGroup.Id,
				"service": inviteTokenService,
			},
		)
		if err == nil && len(tokens) > 0 {
			alreadyToken++
			pending = append(pending, entry)
			continue
		}

		if _, err := GenerateGroupInvite(app, user, defaultGroup); err != nil {
			log.Printf("[backfill] user=%s invite generation failed (non-blocking): %v", user.Id, err)
			failed++
			pending = append(pending, entry)
			continue
		}
		created++
		pending = append(pending, entry)
	}

	log.Printf(
		"[backfill] default group invites complete: group=%s total=%d already_member=%d already_token=%d created=%d failed=%d",
		defaultGroup.Id, total, skippedMember, alreadyToken, created, failed,
	)

	if len(pending) == 0 {
		return
	}

	sort.Slice(pending, func(i, j int) bool {
		if pending[i].DaysSince != pending[j].DaysSince {
			return pending[i].DaysSince > pending[j].DaysSince
		}
		return strings.ToLower(pending[i].FullName) < strings.ToLower(pending[j].FullName)
	})

	email := copyadmin.BuildPendingDefaultGroupEmail(defaultGroup.GetString("name"), pending)
	if !backendinternal.SendAdminFailureEmail(app, email.Subject, email.Body) {
		log.Printf("[backfill] admin email not sent (recipient not configured or send failed)")
	}
}

func StartDefaultGroupInvitesSchedule(app *pocketbase.PocketBase) {
	now := time.Now()
	lastSentYear, lastSentWeek := 0, 0
	if isSundayEvening(now) {
		lastSentYear, lastSentWeek = now.ISOWeek()
	}

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			current := time.Now()
			if !isSundayEvening(current) {
				continue
			}
			year, week := current.ISOWeek()
			if year == lastSentYear && week == lastSentWeek {
				continue
			}
			log.Printf("[backfill] running weekly sunday-evening default-group scan")
			BackfillDefaultGroupInvites(app)
			lastSentYear, lastSentWeek = year, week
		}
	}()
}

func isSundayEvening(now time.Time) bool {
	return now.Weekday() == time.Sunday && now.Hour() >= sundayEveningStartHour
}

func resolveLocalGroupName(app *pocketbase.PocketBase, userID string) string {
	memberships, err := app.FindRecordsByFilter(
		"user_groups",
		"user = {:user}",
		"",
		0,
		0,
		map[string]any{"user": userID},
	)
	if err != nil {
		return ""
	}
	for _, m := range memberships {
		groupID := strings.TrimSpace(m.GetString("group"))
		if groupID == "" {
			continue
		}
		group, err := app.FindRecordById("groups", groupID)
		if err != nil || group == nil {
			continue
		}
		if strings.TrimSpace(group.GetString("type")) == "local" {
			return strings.TrimSpace(group.GetString("name"))
		}
	}
	return ""
}

func daysSince(t time.Time, now time.Time) int {
	if t.IsZero() {
		return 0
	}
	diff := now.Sub(t).Hours() / 24
	if diff < 0 {
		return 0
	}
	return int(math.Floor(diff))
}
