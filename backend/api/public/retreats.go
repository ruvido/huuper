package public

import (
	"net/http"

	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	errInvalidRetreat = "invalid_retreat"
	errRetreatClosed  = "retreat_closed"
	errGuestFields    = "missing_guest_fields"
	errRetreatFull    = "retreat_full"
)

// RetreatDetailsHandler returns public-safe details for a retreat by slug.
// Unlike EventDetailsHandler, it NEVER 404s based on `active` — the public
// page always renders (with a "registrations not open yet" state client
// side when active is false).
func RetreatDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidRetreat, nil)
		}

		retreat, err := retreatsinternal.FindBySlug(app, slug)
		if err != nil || retreat == nil {
			return apis.NewNotFoundError(errInvalidRetreat, err)
		}

		item := retreatsinternal.MapItem(retreat)
		remaining, limited, _ := retreatsinternal.RemainingCapacity(app, retreat)

		// How many seats are left is organiser-side. The page only needs to
		// know whether it can still take registrations.
		response := map[string]any{
			"retreat": item,
			"full":    limited && remaining <= 0,
		}

		if next, err := retreatsinternal.NextRetreat(app, retreat.Id); err == nil && next != nil {
			nextItem := retreatsinternal.MapItem(next)
			response["next_retreat"] = nextItem
		}

		return e.JSON(http.StatusOK, response)
	}
}
