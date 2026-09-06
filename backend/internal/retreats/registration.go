package retreats

import (
	"encoding/json"
	"fmt"
	"strings"

	backendinternal "members/backend/internal"
	paymentsinternal "members/backend/internal/payments"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/core/validators"
)

const maxRegistrationDataBytes = 4000

// RegisterInput describes a new registration submission (member or guest —
// distinguished by whether `User` is set).
type RegisterInput struct {
	Email string
	User  *core.Record
	// SkipApproval is what actually fast-tracks a registration. It is kept
	// separate from User because an authenticated visitor whose account is not
	// approved is still linked to their user record, but must not skip the
	// admin's review.
	SkipApproval bool
	Data         map[string]any
}

// Register creates a pending registration for the retreat. Callers must
// have already checked retreat.active and remaining capacity. Members
// (User != nil) skip the pending-review step entirely: they go straight to
// Activate/MarkAwaitingPayment. Guests (User == nil) stay `pending` until an
// admin approves or rejects them.
func Register(app *pocketbase.PocketBase, retreat *core.Record, in RegisterInput) (*core.Record, string, error) {
	if retreat == nil {
		return nil, "", fmt.Errorf("missing retreat")
	}
	email, err := backendinternal.NormalizeEmail(in.Email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email")
	}

	data := in.Data
	if data == nil {
		data = map[string]any{}
	}
	if in.User != nil {
		// Keep required field non-empty without duplicating profile data.
		data = map[string]any{"linked_user": true}
	}
	if !isRegistrationDataSizeOK(data) {
		return nil, "", fmt.Errorf("registration data too large")
	}

	collection, err := app.FindCollectionByNameOrId("retreat_registrations")
	if err != nil {
		return nil, "", err
	}

	acceptToken, err := GenerateAcceptToken(app)
	if err != nil {
		return nil, "", err
	}

	record := core.NewRecord(collection)
	record.Set("retreat", retreat.Id)
	record.Set("email", email)
	if in.User != nil {
		record.Set("user", in.User.Id)
	}
	record.Set("status", "pending")
	record.Set("accept_token", acceptToken)
	record.Set("accept_expires_at", AcceptTokenExpiry())
	record.Set("data", data)

	if err := app.Save(record); err != nil {
		if isUniqueRegistrationConstraintError(err) {
			return nil, "", fmt.Errorf("already_submitted")
		}
		return nil, "", err
	}

	checkoutURL := ""
	if in.SkipApproval {
		// Approved members skip approval entirely. No payment-link email
		// either: the browser is about to land them on that very page.
		checkoutURL, err = resolveActivation(app, retreat, record, false)
		if err != nil {
			return nil, "", err
		}
	}

	// "We received your request, we'll call you before confirming" is the
	// guest flow talking. Nobody calls a member back and nothing is pending
	// review for them, so the only email their registration produces is the
	// confirmation once the deposit is in.
	if !in.SkipApproval {
		SendRegistrationEmail(app, retreat, email)
		SendAdminNewRegistrationNotification(app, retreat, record, email)
	}

	return record, checkoutURL, nil
}

// Approve moves a pending guest registration forward: straight to `active`
// if the retreat has no deposit configured, otherwise to
// `awaiting_payment` with a Stripe checkout link emailed to the registrant.
// Returns the checkout URL, if any.
func Approve(app *pocketbase.PocketBase, registration *core.Record) (string, error) {
	if registration == nil {
		return "", fmt.Errorf("missing registration")
	}
	if registration.GetString("status") == "active" {
		return "", nil
	}
	retreat, err := app.FindRecordById("retreats", registration.GetString("retreat"))
	if err != nil || retreat == nil {
		return "", fmt.Errorf("retreat not found")
	}
	// The guest is not on the site — they are being approved hours or days
	// after applying — so the payment link has to reach them by email.
	return resolveActivation(app, retreat, registration, true)
}

// emailPaymentLink tells apart the two ways a registration reaches the
// deposit: a member clicking through right now (redirected to Stripe, an
// email would be noise) and a guest approved later (the email is the only
// way they hear about it).
func resolveActivation(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record, emailPaymentLink bool) (string, error) {
	depositCents := DepositCentsForRetreat(retreat)
	if depositCents > 0 && registration.GetString("status") != "awaiting_payment" {
		_, url, err := paymentsinternal.CreateCheckoutSession(app, paymentsinternal.CheckoutInput{
			PurposeType: "retreat_registration",
			PurposeID:   registration.Id,
			Email:       registration.GetString("email"),
			AmountCents: int64(depositCents),
			Currency:    "eur",
			ProductName: strings.TrimSpace(retreat.GetString("title")) + " - deposit",
			SuccessURL:  PaymentSuccessURL(app, retreat),
			CancelURL:   PaymentCancelURL(app, retreat),
		})
		if err != nil {
			return "", err
		}
		if err := MarkAwaitingPayment(app, registration, url, emailPaymentLink); err != nil {
			return "", err
		}
		return url, nil
	}

	return "", Activate(app, registration)
}

// Activate marks a registration active, generates a Telegram invite link
// (if the retreat has data.telegram_invite set) and emails the confirmation.
func Activate(app *pocketbase.PocketBase, registration *core.Record) error {
	if registration == nil {
		return fmt.Errorf("missing registration")
	}
	registration.Set("status", "active")
	if err := app.Save(registration); err != nil {
		return err
	}

	retreat, _ := app.FindRecordById("retreats", registration.GetString("retreat"))
	inviteLink := InviteLinkForRetreat(app, retreat)

	SendAcceptedEmail(app, retreat, registration.GetString("email"), inviteLink)
	return nil
}

// Reject marks a pending guest registration rejected with an admin-supplied
// note.
func Reject(app *pocketbase.PocketBase, registration *core.Record, note string) error {
	if registration == nil {
		return fmt.Errorf("missing registration")
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return fmt.Errorf("missing note")
	}
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	data["rejected"] = note
	registration.Set("data", data)
	registration.Set("status", "rejected")
	return app.Save(registration)
}

// CancelRegistration marks an active/pending registration cancelled with an
// admin-supplied note, freeing up its capacity slot.
func CancelRegistration(app *pocketbase.PocketBase, registration *core.Record, note string) error {
	if registration == nil {
		return fmt.Errorf("missing registration")
	}
	if registration.GetString("status") == "cancelled" {
		return nil
	}
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	if strings.TrimSpace(note) != "" {
		data["cancelled"] = strings.TrimSpace(note)
	}
	registration.Set("data", data)
	registration.Set("status", "cancelled")
	return app.Save(registration)
}

func isRegistrationDataSizeOK(data map[string]any) bool {
	if data == nil {
		return true
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return false
	}
	return len(raw) <= maxRegistrationDataBytes
}

func isUniqueRegistrationConstraintError(err error) bool {
	if err == nil {
		return false
	}
	normalized := validators.NormalizeUniqueIndexError(
		err,
		"retreat_registrations",
		[]string{"retreat", "email"},
	)
	fieldErrors, ok := normalized.(validation.Errors)
	if !ok {
		return false
	}
	_, hasRetreat := fieldErrors["retreat"]
	_, hasEmail := fieldErrors["email"]
	return hasRetreat || hasEmail
}
