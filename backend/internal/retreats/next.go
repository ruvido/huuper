package retreats

import (
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// NextRetreat returns the nearest upcoming, active retreat by start_date,
// excluding the given retreat ID. Used for the public detail page's
// "next occurrence" footer link — a pointer to whatever the next distinct
// retreat is, not a recurrence of the current one (retreats never recur).
func NextRetreat(app *pocketbase.PocketBase, excludeID string) (*core.Record, error) {
	now := time.Now()
	filter := "active = true && start_date >= {:now}"
	params := map[string]any{"now": now}
	if strings.TrimSpace(excludeID) != "" {
		filter += " && id != {:excludeID}"
		params["excludeID"] = strings.TrimSpace(excludeID)
	}
	records, err := app.FindRecordsByFilter("retreats", filter, "start_date", 1, 0, params)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}
