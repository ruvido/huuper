package events

import (
	"fmt"
	"strings"
	"time"

	backendinternal "members/backend/internal"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func BuildAcceptURL(app *pocketbase.PocketBase, token string) string {
	base := strings.TrimRight(app.Settings().Meta.AppURL, "/")
	if base == "" || token == "" {
		return ""
	}
	return base + "/#/event-accept?token=" + token
}

func GenerateAcceptToken(app *pocketbase.PocketBase) (string, error) {
	const attempts = 5
	for i := 0; i < attempts; i++ {
		token := backendinternal.RandomToken()
		if token == "" {
			continue
		}
		unique, err := isTokenUnique(app, token)
		if err != nil {
			return "", err
		}
		if unique {
			return token, nil
		}
	}

	return "", fmt.Errorf("unable to generate unique accept token")
}

func AcceptTokenExpiryForEvent(event *core.Record) time.Time {
	if event == nil {
		return time.Now().UTC()
	}
	eventDate := event.GetDateTime("event_date")
	if eventDate.IsZero() {
		return time.Now().UTC()
	}
	localEventDay := eventDate.Time().In(time.Local)
	endOfDayLocal := time.Date(
		localEventDay.Year(),
		localEventDay.Month(),
		localEventDay.Day(),
		23, 59, 59, 0,
		time.Local,
	)
	return endOfDayLocal.UTC()
}

func isTokenUnique(app *pocketbase.PocketBase, token string) (bool, error) {
	records, err := app.FindRecordsByFilter(
		"event_registrations",
		"accept_token = {:token}",
		"",
		1,
		0,
		map[string]any{"token": token},
	)
	if err != nil {
		return false, err
	}
	return len(records) == 0, nil
}
