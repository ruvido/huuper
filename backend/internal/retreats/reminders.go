package retreats

import (
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	backendinternal "members/backend/internal"
)

// How long to wait before nudging someone who owes the deposit, and how many
// times to do it. These are the defaults; the real values are settings, because
// "three days" is a decision about how insistent to be with people, not a fact
// about the code.
const (
	defaultReminderAfterDays = 3
	defaultReminderMax       = 2
)

// settingsName is the record the retreat module reads its own configuration from.
const settingsName = "retreats"

// ReminderConfig is how the deposit chase is set up.
type ReminderConfig struct {
	AfterDays int // days of silence before a reminder goes out
	Max       int // reminders after the first automatic email (so Max=2 means 3 in all)
}

// LoadReminderConfig reads settings.retreats, falling back to the defaults for
// anything missing so a fresh instance still behaves sensibly.
func LoadReminderConfig(app *pocketbase.PocketBase) ReminderConfig {
	config := ReminderConfig{AfterDays: defaultReminderAfterDays, Max: defaultReminderMax}
	data, err := backendinternal.FindSettingData(app, settingsName)
	if err != nil {
		return config
	}
	if v := intFrom(data["payment_reminder_days"]); v > 0 {
		config.AfterDays = v
	}
	if v := intFrom(data["payment_reminder_max"]); v >= 0 {
		config.Max = v
	}
	return config
}

func intFrom(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return -1
}

// SendPaymentReminders chases everyone who is still to pay the deposit for an
// open retreat. A reminder carries a FRESH checkout link: Stripe expires an
// unused session on its own, so resending the old address would send people to
// a dead page.
func SendPaymentReminders(app *pocketbase.PocketBase) {
	config := LoadReminderConfig(app)
	if config.Max <= 0 || config.AfterDays <= 0 {
		return
	}
	retreatRecords, err := app.FindRecordsByFilter("retreats", "active = true", "start_date", 0, 0)
	if err != nil {
		app.Logger().Warn("retreats: reminder lookup failed", "error", err)
		return
	}
	for _, retreat := range retreatRecords {
		registrations, err := RegistrationsByStatus(app, retreat.Id, "awaiting_payment")
		if err != nil {
			app.Logger().Warn("retreats: reminder registrations failed", "error", err, "retreat", retreat.Id)
			continue
		}
		for _, registration := range registrations {
			sendPaymentReminder(app, retreat, registration, config)
		}
	}
}

// sendPaymentReminder sends one reminder if this registration is due for one.
func sendPaymentReminder(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record, config ReminderConfig) {
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	sent := intFrom(data["payment_reminders_sent"])
	if sent < 0 {
		sent = 0
	}
	if sent >= config.Max {
		return
	}
	// Silence is measured from the last time we wrote to them — the first email
	// if none has been sent yet.
	last := registration.GetDateTime("updated").Time()
	if raw, ok := data["payment_reminder_at"].(string); ok && strings.TrimSpace(raw) != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			last = parsed
		}
	}
	if time.Since(last) < time.Duration(config.AfterDays)*24*time.Hour {
		return
	}

	url, err := ResumeCheckout(app, retreat, registration)
	if err != nil {
		app.Logger().Warn("retreats: reminder checkout failed", "error", err, "registration", registration.Id)
		return
	}
	if !SendPaymentLinkEmail(app, retreat, registration.GetString("email"), url) {
		return
	}

	// Re-read: ResumeCheckout saved the new link on the record.
	data = backendinternal.ParseJSONMap(registration.Get("data"))
	data["payment_reminders_sent"] = sent + 1
	data["payment_reminder_at"] = time.Now().UTC().Format(time.RFC3339)
	registration.Set("data", data)
	if err := app.Save(registration); err != nil {
		app.Logger().Warn("retreats: reminder bookkeeping failed", "error", err, "registration", registration.Id)
	}
}

// StartPaymentRemindersSchedule chases deposits once a day, in the same morning
// slot as the figures: the organiser's day starts with the email and the people
// who owe money hear from us at a civil hour, not at whatever time the container
// happened to restart.
func StartPaymentRemindersSchedule(app *pocketbase.PocketBase) {
	startMorningSchedule("checking deposit reminders", func() { SendPaymentReminders(app) })
}
