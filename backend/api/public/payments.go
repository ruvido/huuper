package public

import (
	"io"
	"net/http"

	paymentsinternal "members/backend/internal/payments"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// StripeWebhookHandler receives Stripe event deliveries. Signature
// verification happens inside payments.HandleWebhook against the
// webhook_secret configured in the "stripe" settings record.
func StripeWebhookHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		payload, err := io.ReadAll(e.Request.Body)
		if err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		signature := e.Request.Header.Get("Stripe-Signature")

		if err := paymentsinternal.HandleWebhook(app, payload, signature); err != nil {
			app.Logger().Warn("stripe webhook failed", "error", err)
			return apis.NewBadRequestError("webhook_error", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"received": true})
	}
}
