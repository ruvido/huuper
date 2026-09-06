package public

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// Outcomes of the approve-from-email click. Must stay in sync with the
// keys under retreats.public.accept in frontend/skeleton/copy/retreats.js.
const (
	acceptStatusApproved = "approved"
	acceptStatusActive   = "active"
	acceptStatusAlready  = "already"
	acceptStatusInvalid  = "invalid"
	acceptStatusFailed   = "failed"
)

type retreatRegistrationPayload struct {
	Email string         `json:"email"`
	Data  map[string]any `json:"data"`
}

// RegisterRetreatHandler creates a registration for a retreat by slug.
// Unlike the detail endpoint, this DOES gate on `active` and capacity.
func RegisterRetreatHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidRetreat, nil)
		}

		retreat, err := retreatsinternal.FindBySlug(app, slug)
		if err != nil || retreat == nil {
			return apis.NewNotFoundError(errInvalidRetreat, err)
		}
		if !retreat.GetBool("active") {
			return apis.NewForbiddenError(errRetreatClosed, nil)
		}

		remaining, limited, err := retreatsinternal.RemainingCapacity(app, retreat)
		if err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}
		if limited && remaining <= 0 {
			return apis.NewForbiddenError(errRetreatFull, nil)
		}

		var payload retreatRegistrationPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}

		// A user gets their password and `approved` in the same step (see
		// requests.Promote), so anyone authenticated here is a member: the only
		// status that disqualifies them is `cancelled`, which an admin can set
		// afterwards without invalidating the password. Same test the rest of
		// the codebase uses (telegram.invites, admin.users) — checking for
		// `approved` instead would wrongly demote the other valid statuses.
		authUser := e.Auth
		canSkipApproval := authUser != nil && strings.TrimSpace(authUser.GetString("status")) != "cancelled"

		linkedUser := authUser
		email := payload.Email
		if linkedUser != nil {
			email = linkedUser.GetString("email")
		} else {
			normalized, err := backendinternal.NormalizeEmail(payload.Email)
			if err != nil {
				return apis.NewBadRequestError(errInvalidEmail, nil)
			}
			email = normalized
			user, err := app.FindFirstRecordByFilter(
				"users",
				"email = {:email}",
				map[string]any{"email": normalized},
			)
			if err == nil && user != nil && strings.TrimSpace(user.GetString("status")) == "approved" {
				linkedUser = user
				canSkipApproval = true
			}
		}

		// Guests submit a pre-registration: the organiser has nothing on file
		// about them and has to call back, so these are required server-side
		// too and not just in the browser. Members skip this — Register()
		// discards submitted data for them and links the user record instead.
		if authUser == nil && linkedUser == nil && !hasRequiredGuestFields(payload.Data) {
			return apis.NewBadRequestError(errGuestFields, nil)
		}

		existing, err := retreatsinternal.FindRegistrationByEmail(app, retreat.Id, email, false)
		if err == nil && existing != nil {
			// Someone who closed the Stripe page has a registration that is
			// only waiting to be paid. Sending them away with "you already
			// signed up" leaves them stuck: they cannot pay, and the unique
			// index refuses a second registration. Hand them a new checkout
			// link instead — the frontend redirects on checkout_url exactly as
			// it does for a first registration.
			if existing.GetString("status") == "awaiting_payment" {
				checkoutURL, err := retreatsinternal.ResumeCheckout(app, retreat, existing)
				if err != nil {
					return apis.NewBadRequestError(errGeneric, err)
				}
				return e.JSON(http.StatusOK, map[string]any{
					"id":           existing.Id,
					"checkout_url": checkoutURL,
				})
			}
			return apis.NewBadRequestError(errAlreadySubmitted, nil)
		}

		record, checkoutURL, err := retreatsinternal.Register(app, retreat, retreatsinternal.RegisterInput{
			Email:        email,
			User:         linkedUser,
			SkipApproval: canSkipApproval,
			Data:         payload.Data,
		})
		if err != nil {
			if err.Error() == "already_submitted" {
				return apis.NewBadRequestError(errAlreadySubmitted, err)
			}
			return apis.NewBadRequestError(errGeneric, err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":           record.Id,
			"checkout_url": checkoutURL,
		})
	}
}

// requiredGuestFields must stay in sync with GUEST_FIELDS in
// frontend/skeleton/assets/js/public/retreat.js.
var requiredGuestFields = []string{"full_name", "birth_year", "provenance", "mobile"}

func hasRequiredGuestFields(data map[string]any) bool {
	if data == nil {
		return false
	}
	for _, key := range requiredGuestFields {
		value, ok := data[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

// AcceptRetreatRegistrationHandler approves a pending registration from the
// link in the admin notification email, so it can be done from a phone
// without the admin panel. Every outcome redirects to /retreat-accept/: the
// only reader here is a person who just tapped a link in their inbox, and a
// JSON body would tell them nothing. The token is time limited; an unknown
// one gets the same answer as an expired one.
func AcceptRetreatRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		result := func(retreat *core.Record, status string) error {
			return e.Redirect(http.StatusFound, retreatsinternal.AcceptResultURL(app, retreat, status))
		}

		registration := retreatsinternal.FindByAcceptToken(app, e.Request.URL.Query().Get("token"))
		if registration == nil {
			return result(nil, acceptStatusInvalid)
		}
		retreat, _ := app.FindRecordById("retreats", registration.GetString("retreat"))

		if registration.GetString("status") != "pending" {
			return result(retreat, acceptStatusAlready)
		}

		if _, err := retreatsinternal.Approve(app, registration); err != nil {
			return result(retreat, acceptStatusFailed)
		}

		// Approve() updates the record in place: with a deposit configured the
		// registrant still has to pay, without one they are already in.
		if registration.GetString("status") == "active" {
			return result(retreat, acceptStatusActive)
		}
		return result(retreat, acceptStatusApproved)
	}
}
