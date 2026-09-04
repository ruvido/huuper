package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

// HandleWebhook verifies and processes a Stripe webhook delivery. On
// checkout.session.completed it marks the matching payment record paid and
// dispatches to the ConfirmHandler registered for its purpose_type.
func HandleWebhook(app *pocketbase.PocketBase, payload []byte, signatureHeader string) error {
	cfg, err := LoadConfig(app)
	if err != nil {
		return err
	}

	event, err := webhook.ConstructEvent(payload, signatureHeader, cfg.WebhookSecret)
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	if event.Type != stripe.EventTypeCheckoutSessionCompleted {
		return nil
	}

	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("decode checkout session: %w", err)
	}

	record, err := app.FindFirstRecordByFilter(
		"payments",
		"stripe_session_id = {:id}",
		map[string]any{"id": session.ID},
	)
	if err != nil || record == nil {
		return fmt.Errorf("payment record not found for session %s", session.ID)
	}

	if record.GetString("status") == "paid" {
		return nil
	}

	record.Set("status", "paid")
	if session.PaymentIntent != nil {
		record.Set("stripe_payment_intent", session.PaymentIntent.ID)
	}
	if err := app.Save(record); err != nil {
		return err
	}

	purposeType := strings.TrimSpace(record.GetString("purpose_type"))
	purposeID := strings.TrimSpace(record.GetString("purpose_id"))
	handler, ok := confirmHandlers[purposeType]
	if !ok {
		log.Printf("payments: no confirm handler registered for purpose_type %q", purposeType)
		return nil
	}
	return handler(app, purposeID)
}
