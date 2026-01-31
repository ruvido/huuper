package migrations

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		if registrations.Fields.GetByName("accepted") == nil {
			registrations.Fields.Add(&core.BoolField{
				Name:     "accepted",
				Required: false,
			})
		}

		if registrations.Fields.GetByName("accept_token") == nil {
			registrations.Fields.Add(&core.TextField{
				Name:     "accept_token",
				Required: false,
				Max:      255,
			})
		}

		if err := app.Save(registrations); err != nil {
			return err
		}

		// Ensure existing records have accepted=false and a token if missing.
		records, err := app.FindRecordsByFilter("event_registrations", "", "", 0, 0)
		if err != nil {
			return err
		}
		for _, record := range records {
			if record.Get("accepted") == nil {
				record.Set("accepted", false)
			}
			if record.GetString("accept_token") == "" {
				record.Set("accept_token", randomToken())
			}
			if err := app.Save(record); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// No downgrade
		return nil
	})
}

func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
