package admin

import (
	"net/http"

	backendinternal "members/backend/internal"
	backendrequests "members/backend/internal/requests"
	backendsettings "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func CreateRequestHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
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

		return e.JSON(http.StatusCreated, map[string]any{
			"id":       record.Id,
			"email":    record.GetString("email"),
			"status":   backendrequests.StatusSubmitted,
			"archived": false,
			"data":     backendinternal.ParseJSONMap(record.Get("data")),
		})
	}
}
