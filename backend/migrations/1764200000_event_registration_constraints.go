package migrations

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("event_registrations", "", "", 0, 0)
		if err != nil {
			return err
		}

		emailKeys := make(map[string]struct{}, len(records))
		tokenKeys := make(map[string]struct{}, len(records))

		for _, record := range records {
			email := strings.ToLower(strings.TrimSpace(record.GetString("email")))
			eventId := strings.TrimSpace(record.GetString("event"))
			if email != "" && eventId != "" {
				key := eventId + "|" + email
				if _, exists := emailKeys[key]; exists {
					return fmt.Errorf("duplicate event registrations detected; resolve duplicates before applying constraints")
				}
				emailKeys[key] = struct{}{}
			}

			token := strings.TrimSpace(record.GetString("accept_token"))
			if token == "" {
				newToken, err := generateUniqueToken(tokenKeys)
				if err != nil {
					return err
				}
				token = newToken
				record.Set("accept_token", token)
				if err := app.Save(record); err != nil {
					return err
				}
				tokenKeys[token] = struct{}{}
				continue
			}

			if _, exists := tokenKeys[token]; exists {
				newToken, err := generateUniqueToken(tokenKeys)
				if err != nil {
					return err
				}
				token = newToken
				record.Set("accept_token", token)
				if err := app.Save(record); err != nil {
					return err
				}
			}
			tokenKeys[token] = struct{}{}
		}

		registrations.AddIndex(
			"idx_event_registrations_event_email",
			true,
			"event, email",
			"email != ''",
		)
		registrations.AddIndex(
			"idx_event_registrations_accept_token",
			true,
			"accept_token",
			"accept_token != ''",
		)

		return app.Save(registrations)
	}, func(app core.App) error {
		registrations, err := app.FindCollectionByNameOrId("event_registrations")
		if err != nil {
			return err
		}

		registrations.RemoveIndex("idx_event_registrations_event_email")
		registrations.RemoveIndex("idx_event_registrations_accept_token")
		return app.Save(registrations)
	})
}

func generateUniqueToken(existing map[string]struct{}) (string, error) {
	for i := 0; i < 5; i++ {
		token := randomToken32()
		if token == "" {
			continue
		}
		if _, exists := existing[token]; !exists {
			return token, nil
		}
	}
	return "", fmt.Errorf("unable to generate unique accept token")
}

func randomToken32() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
