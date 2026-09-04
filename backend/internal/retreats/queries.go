package retreats

import (
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ListItems returns retreats for the given window ("future"/"past"/"all"),
// sorted the same way events.ListForUser sorts events: future ascending
// (soonest first), past descending (most recent first). A retreat with no
// end_date is classified by start_date alone.
func ListItems(app *pocketbase.PocketBase, window string) ([]Item, error) {
	records, err := app.FindRecordsByFilter("retreats", "", "start_date", 0, 0)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	classify := func(record *core.Record) time.Time {
		end := record.GetDateTime("end_date")
		if !end.IsZero() {
			return end.Time()
		}
		return record.GetDateTime("start_date").Time()
	}

	filtered := make([]*core.Record, 0, len(records))
	for _, record := range records {
		isFuture := !classify(record).Before(now)
		switch window {
		case "past":
			if !isFuture {
				filtered = append(filtered, record)
			}
		case "all":
			filtered = append(filtered, record)
		default: // future
			if isFuture {
				filtered = append(filtered, record)
			}
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		ti, tj := classify(filtered[i]), classify(filtered[j])
		if window == "past" {
			return ti.After(tj)
		}
		return ti.Before(tj)
	})

	items := make([]Item, 0, len(filtered))
	for _, record := range filtered {
		items = append(items, MapItem(record))
	}
	return items, nil
}

// ActiveRegistrationFilterSuffix matches registrations that still hold a
// capacity slot (i.e. not rejected/cancelled). Kept separate from "active"
// on purpose: awaiting_payment still reserves a spot until it's resolved.
const ActiveRegistrationFilterSuffix = "status != 'cancelled' && status != 'rejected'"

func FindBySlug(app *pocketbase.PocketBase, slug string) (*core.Record, error) {
	rawSlug := strings.TrimSpace(slug)
	if rawSlug == "" {
		return nil, nil
	}
	return app.FindFirstRecordByFilter(
		"retreats",
		"slug = {:slug}",
		map[string]any{"slug": rawSlug},
	)
}

func FindRegistrationByEmail(app *pocketbase.PocketBase, retreatID string, email string, activeOnly bool) (*core.Record, error) {
	filter := "retreat = {:retreat} && email = {:email}"
	if activeOnly {
		filter += " && " + ActiveRegistrationFilterSuffix
	}
	return app.FindFirstRecordByFilter(
		"retreat_registrations",
		filter,
		map[string]any{
			"retreat": retreatID,
			"email":   email,
		},
	)
}

func FindRegistrationByUser(app *pocketbase.PocketBase, retreatID string, userID string, activeOnly bool) (*core.Record, error) {
	filter := "retreat = {:retreat} && user = {:user}"
	if activeOnly {
		filter += " && " + ActiveRegistrationFilterSuffix
	}
	return app.FindFirstRecordByFilter(
		"retreat_registrations",
		filter,
		map[string]any{
			"retreat": retreatID,
			"user":    userID,
		},
	)
}

// CountReservedRegistrations counts registrations that reserve a capacity
// slot for the retreat (active or awaiting_payment).
func CountReservedRegistrations(app *pocketbase.PocketBase, retreatID string) (int, error) {
	records, err := app.FindRecordsByFilter(
		"retreat_registrations",
		"retreat = {:retreat} && (status = 'active' || status = 'awaiting_payment')",
		"",
		0, 0,
		map[string]any{"retreat": retreatID},
	)
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

// RegistrationsByStatus returns all registrations for a retreat with the
// given status, newest first.
func RegistrationsByStatus(app *pocketbase.PocketBase, retreatID string, status string) ([]*core.Record, error) {
	return app.FindRecordsByFilter(
		"retreat_registrations",
		"retreat = {:retreat} && status = {:status}",
		"-created",
		0, 0,
		map[string]any{"retreat": retreatID, "status": status},
	)
}

// RemainingCapacity returns the number of spots left for the retreat, and
// whether the retreat's data.capacity even defines a limit. When capacity is
// unset (0), registration is treated as uncapped.
func RemainingCapacity(app *pocketbase.PocketBase, retreat *core.Record) (remaining int, limited bool, err error) {
	if retreat == nil {
		return 0, false, nil
	}
	// Hidden column, not `data`: `data` is served verbatim to the public.
	capacity := retreat.GetInt("capacity")
	if capacity <= 0 {
		return 0, false, nil
	}
	reserved, err := CountReservedRegistrations(app, retreat.Id)
	if err != nil {
		return 0, true, err
	}
	remaining = capacity - reserved
	if remaining < 0 {
		remaining = 0
	}
	return remaining, true, nil
}
