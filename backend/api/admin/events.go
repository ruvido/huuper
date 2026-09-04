package admin

import (
	"net/http"
	"strings"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"
	paymentsinternal "members/backend/internal/payments"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type registrationNotePayload struct {
	Note string `json:"note"`
}

func EventDetailsHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventID := e.Request.PathValue("id")
		if eventID == "" {
			return apis.NewBadRequestError("invalid_event", nil)
		}

		event, err := app.FindRecordById("events", eventID)
		if err != nil || event == nil {
			return apis.NewNotFoundError("invalid_event", err)
		}

		attendees, err := eventinternal.ActiveAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		pendingAttendees, err := eventinternal.PendingAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}
		cancelledAttendees, err := eventinternal.CancelledAttendeesForEvent(app, eventID)
		if err != nil {
			return apis.NewBadRequestError("failed_registrations", err)
		}

		item := eventinternal.MapItem(event)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		occurrences, _ := eventinternal.OccurrencesFor(event)
		return e.JSON(http.StatusOK, map[string]any{
			"event":                   item,
			"occurrences":             occurrences,
			"can_edit":                true,
			"registrations":           attendees,
			"pending_registrations":   pendingAttendees,
			"cancelled_registrations": cancelledAttendees,
		})
	}
}

func ApproveRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "active" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_accepted"})
		}

		event, err := app.FindRecordById("events", record.GetString("event"))
		if err != nil || event == nil {
			return apis.NewBadRequestError("invalid_event", err)
		}

		depositCents := eventinternal.DepositCentsForEvent(app, event)
		if depositCents > 0 && record.GetString("status") != "awaiting_payment" {
			_, url, err := paymentsinternal.CreateCheckoutSession(app, paymentsinternal.CheckoutInput{
				PurposeType: "event_registration",
				PurposeID:   record.Id,
				Email:       record.GetString("email"),
				AmountCents: int64(depositCents),
				Currency:    "eur",
				ProductName: strings.TrimSpace(event.GetString("title")) + " - caparra",
				SuccessURL:  eventinternal.PaymentSuccessURL(app),
				CancelURL:   eventinternal.PaymentCancelURL(app),
			})
			if err != nil {
				return apis.NewBadRequestError("failed_checkout", err)
			}
			if err := eventinternal.MarkAwaitingPayment(app, record, url); err != nil {
				return apis.NewBadRequestError("failed_update", err)
			}
			return e.JSON(http.StatusOK, map[string]any{"status": "awaiting_payment"})
		}

		if err := eventinternal.ActivateRegistration(app, record, "events.user.registration_accepted"); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "accepted"})
	}
}

func CancelRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		note := strings.TrimSpace(payload.Note)
		if note == "" {
			return apis.NewBadRequestError("invalid_cancelled_note", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		if record.GetString("status") == "cancelled" {
			return e.JSON(http.StatusOK, map[string]any{"status": "already_cancelled"})
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		data["cancelled"] = note
		record.Set("data", data)
		record.Set("status", "cancelled")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "cancelled"})
	}
}

func RejectRegistrationHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		regID := e.Request.PathValue("id")
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}

		var payload registrationNotePayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}

		note := strings.TrimSpace(payload.Note)
		if note == "" {
			return apis.NewBadRequestError("invalid_rejected_note", nil)
		}

		record, err := app.FindRecordById("event_registrations", regID)
		if err != nil || record == nil {
			return apis.NewNotFoundError("invalid_registration", err)
		}

		data := backendinternal.ParseJSONMap(record.Get("data"))
		data["rejected"] = note
		record.Set("data", data)
		record.Set("status", "rejected")
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}

		return e.JSON(http.StatusOK, map[string]any{"status": "rejected"})
	}
}

func CreateEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		actor, err := backendinternal.RequireAuthenticatedActor(e)
		if err != nil {
			return err
		}
		var input eventinternal.CreateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		record, err := eventinternal.Create(app, actor, input)
		if err != nil {
			return apis.NewBadRequestError("failed_event_create", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusCreated, item)
	}
}

func UpdateEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var input eventinternal.UpdateInput
		if err := e.BindBody(&input); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.Update(app, record, input); err != nil {
			return apis.NewBadRequestError("failed_event_update", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusOK, item)
	}
}

func RescheduleEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload struct {
			EventDate string `json:"event_date"`
			StartDate string `json:"start_date"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		newDate := strings.TrimSpace(payload.StartDate)
		if newDate == "" {
			newDate = payload.EventDate
		}
		if err := eventinternal.Reschedule(app, record, newDate); err != nil {
			return apis.NewBadRequestError("failed_event_reschedule", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusOK, item)
	}
}

func CancelEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		if err := eventinternal.Cancel(app, record); err != nil {
			return apis.NewBadRequestError("failed_event_cancel", err)
		}
		return e.JSON(http.StatusOK, map[string]any{"deleted": 1})
	}
}

func SetEventActiveHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload struct {
			Active *bool `json:"active"`
		}
		if err := e.BindBody(&payload); err != nil || payload.Active == nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		record.Set("active", *payload.Active)
		if err := app.Save(record); err != nil {
			return apis.NewBadRequestError("failed_update", err)
		}
		item := eventinternal.MapItem(record)
		if cfg, err := eventinternal.LoadConfig(app); err == nil {
			eventinternal.ApplyTypeConfig(&item, cfg)
		}
		return e.JSON(http.StatusOK, item)
	}
}

func CancelOccurrenceEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		record, err := loadEventByID(app, e.Request.PathValue("id"))
		if err != nil {
			return err
		}
		var payload struct {
			Date string `json:"date"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if err := eventinternal.CancelOccurrence(app, record, payload.Date); err != nil {
			return apis.NewBadRequestError("failed_occurrence_cancel", err)
		}
		occurrences, _ := eventinternal.OccurrencesFor(record)
		return e.JSON(http.StatusOK, map[string]any{"occurrences": occurrences})
	}
}

func MarkAttendanceHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		regID := strings.TrimSpace(e.Request.PathValue("id"))
		if regID == "" {
			return apis.NewBadRequestError("invalid_registration", nil)
		}
		registration, err := app.FindRecordById("event_registrations", regID)
		if err != nil || registration == nil {
			return apis.NewNotFoundError("registration_not_found", err)
		}
		var payload struct {
			Attended *bool `json:"attended"`
		}
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError("invalid_payload", err)
		}
		if payload.Attended == nil {
			if err := eventinternal.ClearAttendance(app, registration); err != nil {
				return apis.NewBadRequestError("failed_attendance_clear", err)
			}
		} else {
			if err := eventinternal.MarkAttendance(app, registration, *payload.Attended); err != nil {
				return apis.NewBadRequestError("failed_attendance", err)
			}
		}
		return e.JSON(http.StatusOK, map[string]any{"ok": true})
	}
}

func loadEventByID(app *pocketbase.PocketBase, id string) (*core.Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apis.NewBadRequestError("invalid_event", nil)
	}
	record, err := app.FindRecordById("events", id)
	if err != nil || record == nil {
		return nil, apis.NewNotFoundError("event_not_found", err)
	}
	return record, nil
}
