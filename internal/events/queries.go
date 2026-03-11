package events

import (
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const ActiveRegistrationFilterSuffix = "status != 'cancelled' && status != 'rejected'"

func FindBySlug(app *pocketbase.PocketBase, slug string) (*core.Record, error) {
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

func FindRegistrationByUser(app *pocketbase.PocketBase, eventID string, userID string, activeOnly bool) (*core.Record, error) {
	filter := "event = {:event} && user = {:user}"
	if activeOnly {
		filter += " && " + ActiveRegistrationFilterSuffix
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

func FindRegistrationByEmail(app *pocketbase.PocketBase, eventID string, email string, activeOnly bool) (*core.Record, error) {
	filter := "event = {:event} && email = {:email}"
	if activeOnly {
		filter += " && " + ActiveRegistrationFilterSuffix
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
