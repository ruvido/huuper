package public

import (
	"net/http"
	"strings"
	"time"

	backendinternal "members/internal"
	eventinternal "members/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type registrationPayload struct {
	Email string         `json:"email"`
	Data  map[string]any `json:"data"`
}

const (
	errInvalidEvent     = "invalid_event"
	errInvalidEmail     = "invalid_email"
	errEventClosed      = "event_closed"
	errAlreadySubmitted = "already_submitted"
	errGeneric          = "error_generic"
)

// RegisterEventHandler creates a registration for an active event by slug.
func RegisterEventHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		slug := e.Request.PathValue("slug")
		if slug == "" {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		event, err := eventinternal.FindBySlug(app, slug)
		if err != nil {
			return apis.NewNotFoundError(errInvalidEvent, err)
		}

		if !event.GetBool("active") {
			return apis.NewForbiddenError(errEventClosed, nil)
		}

		eventDate := event.GetDateTime("event_date")
		if eventDate.IsZero() {
			return apis.NewBadRequestError(errInvalidEvent, nil)
		}

		now := time.Now().In(time.Local)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		eventDay := eventDate.Time().In(now.Location())
		eventDay = time.Date(eventDay.Year(), eventDay.Month(), eventDay.Day(), 0, 0, 0, 0, eventDay.Location())

		if !eventDay.After(today) {
			return apis.NewForbiddenError(errEventClosed, nil)
		}

		var payload registrationPayload
		if err := e.BindBody(&payload); err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}
		if payload.Data == nil {
			payload.Data = map[string]any{}
		}
		eventinternal.NormalizeRegistrationNames(payload.Data)
		recipient, err := backendinternal.NormalizeEmail(payload.Email)
		if err != nil {
			return apis.NewBadRequestError(errInvalidEmail, nil)
		}

		linkedUser := e.Auth
		if linkedUser == nil {
			user, err := app.FindFirstRecordByFilter(
				"users",
				"email = {:email}",
				map[string]any{"email": recipient},
			)
			if err == nil && user != nil && strings.TrimSpace(user.GetString("status")) == "active" {
				linkedUser = user
			}
		}

		registrationData := payload.Data
		if linkedUser != nil {
			// Keep required field non-empty without duplicating profile data.
			registrationData = map[string]any{"linked_user": true}
		}
		if !eventinternal.IsRegistrationDataSizeOK(registrationData) {
			return apis.NewBadRequestError(errGeneric, nil)
		}

		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return apis.NewNotFoundError(errGeneric, err)
		}

		existing, err := eventinternal.FindRegistrationByEmail(app, event.Id, recipient, false)
		if err == nil && existing != nil {
			return apis.NewBadRequestError(errAlreadySubmitted, nil)
		}

		acceptToken, err := eventinternal.GenerateAcceptToken(app)
		if err != nil {
			return apis.NewBadRequestError(errGeneric, err)
		}

		record := core.NewRecord(registrations)
		record.Set("accept_token", acceptToken)
		record.Set("accept_expires_at", eventinternal.AcceptTokenExpiryForEvent(event))
		record.Set("event", event.Id)
		if linkedUser != nil {
			record.Set("user", linkedUser.Id)
		}
		record.Set("email", recipient)
		record.Set("status", "pending")
		record.Set("data", registrationData)

		if err := app.Save(record); err != nil {
			if eventinternal.IsUniqueRegistrationConstraintError(err) {
				return apis.NewBadRequestError(errAlreadySubmitted, err)
			}
			return apis.NewBadRequestError(errGeneric, err)
		}

		emailSent := false
		effectiveData := registrationData
		if linkedUser != nil {
			effectiveData = backendinternal.ParseJSONMap(linkedUser.Get("data"))
		}
		if linkedUser != nil {
			if err := eventinternal.ActivateRegistration(app, record, eventinternal.TemplateKindUserRegistrationAccepted); err != nil {
				return apis.NewBadRequestError(errGeneric, err)
			}
		}
		emailSent = eventinternal.SendRegistrationEmail(app, event, recipient)
		eventinternal.SendAdminNotification(app, event, recipient, record.GetString("accept_token"), effectiveData)

		return e.JSON(http.StatusCreated, map[string]any{
			"id":         record.Id,
			"email_sent": emailSent,
		})
	}
}
