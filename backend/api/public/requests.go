package public

import (
	"net/http"
	"strings"
	"time"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"
	backendsettings "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type requestEmailOTPPayload struct {
	Email    string `json:"email"`
	OTPToken string `json:"otp_token"`
	OTPCode  string `json:"otp_code"`
}

func RequestEmailOTPHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var payload requestEmailOTPPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		result, err := backendrequests.GenerateRequestEmailOTP(app, payload.Email)
		if err != nil {
			return err
		}
		if !result.Delivery.Accepted() {
			return apis.NewBadRequestError("email_otp_not_sent", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"otp_token": result.Token,
			"expires":   result.Expires,
		})
	}
}

func VerifyRequestEmailOTPHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var payload requestEmailOTPPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := backendrequests.ConfirmRequestEmailOTP(app, payload.Email, payload.OTPToken, payload.OTPCode); err != nil {
			return err
		}
		return e.JSON(http.StatusOK, map[string]any{"verified": true})
	}
}

func SubmitRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		signup, err := backendrequests.LoadSignupSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_signup_settings", err)
		}
		profileSchema, err := backendrequests.LoadProfileSchemaSettings(app)
		if err != nil {
			return apis.NewBadRequestError("invalid_profile_schema", err)
		}
		flowData, err := backendsettings.FindSettingData(app, "requests_flow")
		if err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}
		if _, err := backendrequests.ParseFlowConfig(flowData); err != nil {
			return apis.NewBadRequestError("invalid_requests_flow_settings", err)
		}

		var raw map[string]any
		if err := e.BindBody(&raw); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		input := backendrequests.NormalizeSubmitInput(raw)
		otpToken := strings.TrimSpace(backendinternal.AnyToString(input["otp_token"]))
		otpCode := strings.TrimSpace(backendinternal.AnyToString(input["otp_code"]))
		delete(input, "otp_token")
		delete(input, "otp_code")

		data, email, err := backendrequests.ValidateAndBuildData(input, signup, profileSchema)
		if err != nil {
			return apis.NewBadRequestError("invalid_request_data", err)
		}

		requestsCollection, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return apis.NewNotFoundError("requests_collection_not_found", err)
		}

		data = backendrequests.SetRequestFlowSnapshotData(data, flowData)

		var record *core.Record
		txErr := app.RunInTransaction(func(txApp core.App) error {
			if err := backendrequests.EnsureSubmitEmailAvailableTx(txApp, email); err != nil {
				return err
			}
			if err := backendrequests.VerifyRequestEmailOTP(txApp, email, otpToken, otpCode); err != nil {
				return err
			}
			data = backendrequests.MarkEmailVerified(data, time.Now().UTC())
			r := core.NewRecord(requestsCollection)
			r.Set("email", email)
			r.Set("data", data)
			r.Set("archived", false)
			if err := txApp.Save(r); err != nil {
				if backendrequests.IsDuplicateEmailError(err) {
					return apis.NewBadRequestError("email_exists_request", nil)
				}
				return apis.NewBadRequestError("failed_to_create_request", err)
			}
			record = r
			return nil
		})
		if txErr != nil {
			return txErr
		}

		backendrequests.NotifyNewRequest(app, record, data)
		delivery := backendrequests.NotifyRequestSubmitted(app, record, data)
		notifications := map[string]any{
			"request_submitted_email": map[string]any{
				"accepted": delivery.Accepted(),
				"sent":     delivery.Sent,
				"failed":   delivery.Failed,
			},
		}
		data["__notifications"] = notifications
		record.Set("data", data)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_update_request", err)
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"id":            record.Id,
			"email":         record.GetString("email"),
			"status":        backendrequests.StatusSubmitted,
			"archived":      false,
			"data":          data,
			"notifications": notifications,
		})
	}
}
