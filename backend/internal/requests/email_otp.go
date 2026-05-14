package requests

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	emailOTPTokenService  = "request_email_otp"
	emailOTPTemplateKind  = "requests.email_otp"
	emailOTPTTL           = 10 * time.Minute
	emailOTPMaxAttempts   = 5
	emailOTPCodeDigits    = 4
	emailOTPTokenAttempts = 5
)

type EmailOTPResult struct {
	Token    string
	Expires  time.Time
	Delivery EmailDelivery
}

func MarkEmailVerified(data map[string]any, at time.Time) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	data["__email_verification"] = map[string]any{
		"verified": true,
		"at":       at.UTC().Format(time.RFC3339),
	}
	return data
}

func EmailVerified(data map[string]any) bool {
	raw, ok := data["__email_verification"].(map[string]any)
	if !ok {
		return false
	}
	return raw["verified"] == true
}

func GenerateRequestEmailOTP(app *pocketbase.PocketBase, email string) (EmailOTPResult, error) {
	normalizedEmail, err := backendinternal.NormalizeEmail(email)
	if err != nil {
		return EmailOTPResult{}, apis.NewBadRequestError("invalid_email", nil)
	}
	if err := EnsureSubmitEmailAvailable(app, normalizedEmail); err != nil {
		return EmailOTPResult{}, err
	}
	if err := deleteExpiredEmailOTPTokens(app, time.Now().UTC()); err != nil {
		return EmailOTPResult{}, apis.NewBadRequestError("failed_to_cleanup_email_otp", err)
	}
	recent, err := hasRecentEmailOTP(app, normalizedEmail, time.Now().UTC().Add(-1*time.Minute))
	if err != nil {
		return EmailOTPResult{}, apis.NewBadRequestError("failed_to_check_email_otp", err)
	}
	if recent {
		return EmailOTPResult{}, apis.NewBadRequestError("email_otp_recently_sent", nil)
	}

	tokensCollection, err := app.FindCollectionByNameOrId("tokens")
	if err != nil {
		return EmailOTPResult{}, apis.NewNotFoundError("tokens_collection_not_found", err)
	}

	if err := deleteEmailOTPTokens(app, normalizedEmail); err != nil {
		return EmailOTPResult{}, apis.NewBadRequestError("failed_to_reset_email_otp", err)
	}

	code, err := randomOTPCode(emailOTPCodeDigits)
	if err != nil {
		return EmailOTPResult{}, apis.NewBadRequestError("failed_to_generate_email_otp", err)
	}

	expires := time.Now().UTC().Add(emailOTPTTL)
	var publicToken string
	for i := 0; i < emailOTPTokenAttempts; i++ {
		publicToken = backendinternal.RandomToken()
		if publicToken == "" {
			continue
		}
		record := core.NewRecord(tokensCollection)
		record.Set("token", publicToken)
		record.Set("email", normalizedEmail)
		record.Set("service", emailOTPTokenService)
		record.Set("expires", expires)
		record.Set("attempts", 0)
		record.Set("code", code)
		if err := app.Save(record); err == nil {
			break
		}
		publicToken = ""
	}
	if publicToken == "" {
		return EmailOTPResult{}, fmt.Errorf("unable to generate email otp token")
	}

	delivery := sendEmailOTP(app, normalizedEmail, code)
	if !delivery.Accepted() {
		_ = deleteEmailOTPToken(app, publicToken)
	}
	return EmailOTPResult{Token: publicToken, Expires: expires, Delivery: delivery}, nil
}

func VerifyRequestEmailOTP(app core.App, email string, token string, code string) error {
	record, err := findEmailOTPRecord(app, email, token, code)
	if err != nil {
		return err
	}
	return app.Delete(record)
}

func ConfirmRequestEmailOTP(app core.App, email string, token string, code string) error {
	record, err := findEmailOTPRecord(app, email, token, code)
	if err != nil {
		return err
	}
	record.Set("verified", true)
	return app.Save(record)
}

func findEmailOTPRecord(app core.App, email string, token string, code string) (*core.Record, error) {
	normalizedEmail, err := backendinternal.NormalizeEmail(email)
	if err != nil {
		return nil, apis.NewBadRequestError("invalid_email", nil)
	}

	token = strings.TrimSpace(token)
	code = strings.TrimSpace(code)
	if token == "" || code == "" {
		return nil, apis.NewBadRequestError("missing_email_otp", nil)
	}

	record, err := app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = {:service} && email = {:email}",
		map[string]any{
			"token":   token,
			"service": emailOTPTokenService,
			"email":   normalizedEmail,
		},
	)
	if err != nil || record == nil {
		return nil, apis.NewBadRequestError("invalid_email_otp", err)
	}

	if expires := record.GetDateTime("expires").Time(); expires.IsZero() || !expires.After(time.Now().UTC()) {
		_ = app.Delete(record)
		return nil, apis.NewBadRequestError("expired_email_otp", nil)
	}

	if record.GetBool("verified") {
		return record, nil
	}

	attempts := record.GetInt("attempts")
	if attempts >= emailOTPMaxAttempts {
		_ = app.Delete(record)
		return nil, apis.NewBadRequestError("invalid_email_otp", nil)
	}

	expected := strings.TrimSpace(record.GetString("code"))
	if expected == "" || code != expected {
		record.Set("attempts", attempts+1)
		_ = app.Save(record)
		return nil, apis.NewBadRequestError("invalid_email_otp", nil)
	}

	return record, nil
}

func deleteEmailOTPTokens(app core.App, email string) error {
	records, err := app.FindRecordsByFilter(
		"tokens",
		"email = {:email} && service = {:service}",
		"",
		500,
		0,
		map[string]any{
			"email":   strings.TrimSpace(email),
			"service": emailOTPTokenService,
		},
	)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record != nil {
			_ = app.Delete(record)
		}
	}
	return nil
}

func hasRecentEmailOTP(app core.App, email string, cutoff time.Time) (bool, error) {
	records, err := app.FindRecordsByFilter(
		"tokens",
		"email = {:email} && service = {:service} && created >= {:cutoff}",
		"",
		1,
		0,
		map[string]any{
			"email":   strings.TrimSpace(email),
			"service": emailOTPTokenService,
			"cutoff":  cutoff,
		},
	)
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

func deleteEmailOTPToken(app core.App, token string) error {
	record, err := app.FindFirstRecordByFilter(
		"tokens",
		"token = {:token} && service = {:service}",
		map[string]any{
			"token":   strings.TrimSpace(token),
			"service": emailOTPTokenService,
		},
	)
	if err != nil || record == nil {
		return nil
	}
	return app.Delete(record)
}

func deleteExpiredEmailOTPTokens(app core.App, now time.Time) error {
	records, err := app.FindRecordsByFilter(
		"tokens",
		"service = {:service} && expires <= {:now}",
		"",
		500,
		0,
		map[string]any{
			"service": emailOTPTokenService,
			"now":     now,
		},
	)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record != nil {
			_ = app.Delete(record)
		}
	}
	return nil
}

func sendEmailOTP(app *pocketbase.PocketBase, email string, code string) EmailDelivery {
	template, found, err := loadNotificationTemplate(app, emailOTPTemplateKind)
	if err != nil || !found {
		app.Logger().Warn("Failed to load request email otp template", "kind", emailOTPTemplateKind, "error", err)
		return EmailDelivery{Failed: 1}
	}
	recipient, ok := backendinternal.ParseAddress(email)
	if !ok {
		return EmailDelivery{Failed: 1}
	}
	values := map[string]string{"otp_code": code}
	sent, failed := sendNotificationEmail(app, []mail.Address{recipient}, template, values)
	return EmailDelivery{Sent: sent, Failed: failed}
}

func randomOTPCode(digits int) (string, error) {
	if digits < 1 {
		return "", fmt.Errorf("invalid otp digits")
	}
	max := big.NewInt(10)
	var b strings.Builder
	for i := 0; i < digits; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteString(n.String())
	}
	return b.String(), nil
}
