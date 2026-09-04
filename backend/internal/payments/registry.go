package payments

import "github.com/pocketbase/pocketbase"

// ConfirmHandler is invoked once a payment for the given purpose is confirmed
// paid via the Stripe webhook. purposeID is the record ID within whatever
// collection the purpose_type owns (e.g. an event_registrations ID).
type ConfirmHandler func(app *pocketbase.PocketBase, purposeID string) error

var confirmHandlers = map[string]ConfirmHandler{}

// RegisterConfirmHandler wires a purpose_type (e.g. "event_registration",
// "merch_order") to the callback that activates it once paid. Call once at
// startup from route registration.
func RegisterConfirmHandler(purposeType string, handler ConfirmHandler) {
	confirmHandlers[purposeType] = handler
}
