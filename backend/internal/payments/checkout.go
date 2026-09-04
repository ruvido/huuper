package payments

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type CheckoutInput struct {
	PurposeType string
	PurposeID   string
	Email       string
	AmountCents int64
	Currency    string
	ProductName string
	SuccessURL  string
	CancelURL   string
}

// CreateCheckoutSession opens a Stripe Checkout Session for the given
// purpose (e.g. an event registration deposit) and records it as a pending
// payment. purpose_type/purpose_id let ConfirmPayment route the webhook
// callback back to the right module once paid.
func CreateCheckoutSession(app *pocketbase.PocketBase, in CheckoutInput) (*core.Record, string, error) {
	purposeType := strings.TrimSpace(in.PurposeType)
	purposeID := strings.TrimSpace(in.PurposeID)
	email := strings.TrimSpace(in.Email)
	currency := strings.ToLower(strings.TrimSpace(in.Currency))
	if purposeType == "" || purposeID == "" || email == "" {
		return nil, "", fmt.Errorf("missing purpose_type, purpose_id or email")
	}
	if in.AmountCents <= 0 {
		return nil, "", fmt.Errorf("invalid amount")
	}
	if currency == "" {
		currency = "eur"
	}

	cfg, err := LoadConfig(app)
	if err != nil {
		return nil, "", err
	}
	stripe.Key = cfg.SecretKey

	params := &stripe.CheckoutSessionParams{
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		CustomerEmail: stripe.String(email),
		SuccessURL:    stripe.String(in.SuccessURL),
		CancelURL:     stripe.String(in.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(currency),
					UnitAmount: stripe.Int64(in.AmountCents),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(in.ProductName),
					},
				},
			},
		},
	}
	params.AddMetadata("purpose_type", purposeType)
	params.AddMetadata("purpose_id", purposeID)

	stripeSession, err := session.New(params)
	if err != nil {
		return nil, "", fmt.Errorf("stripe checkout session: %w", err)
	}

	collection, err := app.FindCollectionByNameOrId("payments")
	if err != nil {
		return nil, "", err
	}

	record := core.NewRecord(collection)
	record.Set("purpose_type", purposeType)
	record.Set("purpose_id", purposeID)
	record.Set("email", email)
	record.Set("amount", in.AmountCents)
	record.Set("currency", currency)
	record.Set("status", "pending")
	record.Set("stripe_session_id", stripeSession.ID)
	if err := app.Save(record); err != nil {
		return nil, "", err
	}

	return record, stripeSession.URL, nil
}
