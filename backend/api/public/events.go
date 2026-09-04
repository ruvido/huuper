package public

import (
	"net/http"
	"strings"
	"time"

	eventinternal "members/backend/internal/events"
	paymentsinternal "members/backend/internal/payments"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// EventDetailsHandler returns public-safe details for an active event by
// slug: no attendee lists, no admin fields — just what the public landing
// page needs to render and decide whether to show a deposit.
func EventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		event, err := eventinternal.FindBySlug(app, slug)
		if err != nil || event == nil || !event.GetBool("active") {
			return apis.NewNotFoundError(errInvalidEvent, err)
		}

		item := eventinternal.MapItem(event)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		occurrences, _ := eventinternal.OccurrencesFor(event)

		return e.JSON(http.StatusOK, map[string]any{
			"event":       item,
			"occurrences": occurrences,
		})
	}
}

func AcceptEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := e.Request.URL.Query().Get("token")
		if token == "" {
			return apis.NewBadRequestError("Token mancante", nil)
		}

		record, err := app.FindFirstRecordByFilter(
			"event_registrations",
			"accept_token = {:token}",
			map[string]any{"token": token},
		)
		if err != nil || record == nil {
			return apis.NewNotFoundError("Registrazione non trovata", err)
		}

		expiresAt := record.GetDateTime("accept_expires_at")
		if !expiresAt.IsZero() && time.Now().After(expiresAt.Time()) {
			return apis.NewBadRequestError("token_expired", nil)
		}

		if record.GetString("status") == "active" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
		}

		event, err := app.FindRecordById("events", record.GetString("event"))
		if err != nil || event == nil {
			return apis.NewBadRequestError("Evento non trovato", err)
		}

		depositCents := eventinternal.DepositCentsForEvent(app, event)
		if depositCents > 0 && record.GetString("status") != "awaiting_payment" {
			_, url, err := paymentsinternal.CreateCheckoutSession(app, paymentsinternal.CheckoutInput{
				PurposeType: "event_registration",
				PurposeID:   record.Id,
				Email:       record.GetString("email"),
				AmountCents: int64(depositCents),
				Currency:    "eur",
				ProductName: strings.TrimSpace(event.GetString("title")) + " - caparra",
				SuccessURL:  eventinternal.PaymentSuccessURL(app),
				CancelURL:   eventinternal.PaymentCancelURL(app),
			})
			if err != nil {
				return apis.NewBadRequestError("Checkout fallito", err)
			}
			if err := eventinternal.MarkAwaitingPayment(app, record, url); err != nil {
				return apis.NewBadRequestError("Aggiornamento fallito", err)
			}
			return e.JSON(http.StatusOK, map[string]any{"status": "awaiting_payment"})
		}

		if err := eventinternal.ActivateRegistration(app, record, "events.user.registration_accepted"); err != nil {
			return apis.NewBadRequestError("Aggiornamento fallito", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}
