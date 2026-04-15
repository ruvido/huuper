package public

import (
	"net/http"

	backendrequests "members/backend/internal/requests"
	backendsettings "members/backend/internal/settings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

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
		data, email, err := backendrequests.ValidateAndBuildData(input, signup, profileSchema)
		if err != nil {
			return apis.NewBadRequestError("invalid_request_data", err)
		}
		if err := backendrequests.EnsureSubmitEmailAvailable(app, email); err != nil {
			return err
		}

		requestsCollection, err := app.FindCollectionByNameOrId("requests")
		if err != nil {
			return apis.NewNotFoundError("requests_collection_not_found", err)
		}

		record := core.NewRecord(requestsCollection)
		record.Set("email", email)
		data = backendrequests.SetRequestFlowSnapshotData(data, flowData)
		record.Set("data", data)
		record.Set("rejected", false)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_to_create_request", err)
		}

		backendrequests.NotifyNewRequest(app, record, data)
		backendrequests.NotifyRequestSubmitted(app, record, data)

		return e.JSON(http.StatusCreated, map[string]any{
			"id":       record.Id,
			"email":    record.GetString("email"),
			"status":   backendrequests.StatusSubmitted,
			"rejected": false,
			"data":     data,
		})
	}
}
