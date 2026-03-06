package api

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const activeEventRegistrationFilterSuffix = "status != 'cancelled' && status != 'rejected'"

func findEventBySlug(app *pocketbase.PocketBase, slug string) (*core.Record, error) {
	rawSlug := strings.TrimSpace(slug)
	if rawSlug == "" {
		return nil, nil
	}

	return app.FindFirstRecordByFilter(
		"events",
		"slug = {:slug}",
		map[string]any{"slug": rawSlug},
	)
}

func findEventRegistrationByUser(app *pocketbase.PocketBase, eventID string, userID string, activeOnly bool) (*core.Record, error) {
	filter := "event = {:event} && user = {:user}"
	if activeOnly {
		filter += " && " + activeEventRegistrationFilterSuffix
	}
	return app.FindFirstRecordByFilter(
		"event_registrations",
		filter,
		map[string]any{
			"event": eventID,
			"user":  userID,
		},
	)
}

func findEventRegistrationByEmail(app *pocketbase.PocketBase, eventID string, email string, activeOnly bool) (*core.Record, error) {
	filter := "event = {:event} && email = {:email}"
	if activeOnly {
		filter += " && " + activeEventRegistrationFilterSuffix
	}
	return app.FindFirstRecordByFilter(
		"event_registrations",
		filter,
		map[string]any{
			"event": eventID,
			"email": email,
		},
	)
}
