package admin

import (
	"net/http"
	"strings"

	retreatsinternal "members/backend/internal/retreats"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func RetreatsListHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		records, err := app.FindRecordsByFilter("retreats", "", "-start_date", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_retreats", err)
		}
		items := make([]retreatsinternal.Item, 0, len(records))
		for _, record := range records {
			items = append(items, retreatsinternal.MapItem(record))
		}
		return e.JSON(http.StatusOK, map[string]any{"items": items})
	}
}

func RetreatCreateHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var input retreatsinternal.CreateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		record, err := retreatsinternal.Create(app, input)
		if err != nil {
			return apis.NewBadRequestError("failed_retreat_create", err)
		}
		return e.JSON(http.StatusCreated, retreatsinternal.MapItem(record))
	}
}

func RetreatUpdateHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var input retreatsinternal.UpdateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := retreatsinternal.Update(app, record, input); err != nil {
			return apis.NewBadRequestError("failed_retreat_update", err)
		}
		return e.JSON(http.StatusOK, retreatsinternal.MapItem(record))
	}
}

func RetreatCancelHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		if err := retreatsinternal.Cancel(app, record); err != nil {
			return apis.NewBadRequestError("failed_retreat_cancel", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"deleted": 1})
	}
}

func RetreatDetailsAdminHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}

		pending, err := retreatsinternal.RegistrationsByStatus(app, record.Id, "pending")
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		active, err := retreatsinternal.RegistrationsByStatus(app, record.Id, "active")
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		awaitingPayment, err := retreatsinternal.RegistrationsByStatus(app, record.Id, "awaiting_payment")
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		rejected, err := retreatsinternal.RegistrationsByStatus(app, record.Id, "rejected")
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		cancelled, err := retreatsinternal.RegistrationsByStatus(app, record.Id, "cancelled")
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		remaining, limited, _ := retreatsinternal.RemainingCapacity(app, record)

		response := map[string]any{
			"retreat":                        retreatsinternal.MapItem(record),
			"pending_registrations":          mapRegistrations(pending),
			"active_registrations":           mapRegistrations(active),
			"awaiting_payment_registrations": mapRegistrations(awaitingPayment),
			"rejected_registrations":         mapRegistrations(rejected),
			"cancelled_registrations":        mapRegistrations(cancelled),
		}
		if limited {
			response["spots_remaining"] = remaining
		}

		return e.JSON(http.StatusOK, response)
	}
}

func RetreatUploadGalleryHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		files, err := e.FindUploadedFiles("gallery")
		if err != nil {
			return apis.NewBadRequestError("invalid_gallery_files", err)
		}
		if err := retreatsinternal.AppendGalleryFiles(app, record, files); err != nil {
			return apis.NewBadRequestError("failed_gallery_upload", err)
		}
		return e.JSON(http.StatusOK, retreatsinternal.MapItem(record))
	}
}

func RetreatRemoveGalleryFileHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		filename := e.Request.PathValue("filename")
		if err := retreatsinternal.RemoveGalleryFile(app, record, filename); err != nil {
			return apis.NewBadRequestError("failed_gallery_remove", err)
		}
		return e.JSON(http.StatusOK, retreatsinternal.MapItem(record))
	}
}

func RetreatRegistrationApproveHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		registration, err := loadRetreatRegistrationByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		checkoutURL, err := retreatsinternal.Approve(app, registration)
		if err != nil {
			return apis.NewBadRequestError("failed_approve", err)
		}
		status := "accepted"
		if checkoutURL != "" {
			status = "awaiting_payment"
		}
		return e.JSON(http.StatusOK, map[string]any{"status": status, "checkout_url": checkoutURL})
	}
}

func RetreatRegistrationRejectHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		registration, err := loadRetreatRegistrationByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := retreatsinternal.Reject(app, registration, payload.Note); err != nil {
			return apis.NewBadRequestError("failed_reject", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"status": "rejected"})
	}
}

func RetreatRegistrationCancelHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		registration, err := loadRetreatRegistrationByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := retreatsinternal.CancelRegistration(app, registration, payload.Note); err != nil {
			return apis.NewBadRequestError("failed_cancel", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"status": "cancelled"})
	}
}

type broadcastEmailPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func RetreatBroadcastEmailHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadRetreatByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload broadcastEmailPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		subject := strings.TrimSpace(payload.Subject)
		body := strings.TrimSpace(payload.Body)
		if subject == "" || body == "" {
			return apis.NewBadRequestError("invalid_broadcast_email", nil)
		}
		sent, failed, err := retreatsinternal.BroadcastEmail(app, record, subject, body)
		if err != nil {
			return apis.NewBadRequestError("failed_broadcast", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"sent": sent, "failed": failed})
	}
}

func mapRegistrations(records []*core.Record) []retreatsinternal.RegistrationItem {
	items := make([]retreatsinternal.RegistrationItem, 0, len(records))
	for _, record := range records {
		items = append(items, retreatsinternal.MapRegistrationItem(record))
	}
	return items
}

func loadRetreatByID(app *pocketbase.PocketBase, id string) (*core.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_retreat", nil)
	}
	record, err := app.FindRecordById("retreats", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("retreat_not_found", err)
	}
	return record, nil
}

func loadRetreatRegistrationByID(app *pocketbase.PocketBase, id string) (*core.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_registration", nil)
	}
	record, err := app.FindRecordById("retreat_registrations", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("registration_not_found", err)
	}
	return record, nil
}
