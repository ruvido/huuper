package events

import (
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func MarkAttendance(app *pocketbase.PocketBase, registration *core.Record, attended bool) error {
	if registration == nil {
		return fmt.Errorf("missing registration")
	}
	eventID := strings.TrimSpace(registration.GetString("event"))
	if eventID == "" {
		return fmt.Errorf("registration has no event")
	}
	event, err := app.FindRecordById("events", eventID)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}
	if !isPast(event) {
		return fmt.Errorf("attendance can only be marked on past events")
	}
	registration.Set("attended", attended)
	return app.Save(registration)
}

func ClearAttendance(app *pocketbase.PocketBase, registration *core.Record) error {
	if registration == nil {
		return fmt.Errorf("missing registration")
	}
	registration.Set("attended", nil)
	return app.Save(registration)
}
