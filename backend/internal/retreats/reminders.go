package retreats

import (
	"log"
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
		sent := 0
		for _, registration := range registrations {
			if sendPaymentReminder(app, retreat, registration, config) {
				sent++
			}
		}
		// Says what was done and to how many: a chase that quietly does nothing is
		// indistinguishable from one that has nobody to chase.
		log.Printf("[retreats] %s: %d awaiting payment, %d reminded (after %d days, max %d)",
			retreat.GetString("slug"), len(registrations), sent, config.AfterDays, config.Max)
	}
}

// sendPaymentReminder sends one reminder if this registration is due for one.
func sendPaymentReminder(app *pocketbase.PocketBase, retreat *core.Record, registration *core.Record, config ReminderConfig) bool {
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	sent := intFrom(data["payment_reminders_sent"])
	if sent < 0 {
		sent = 0
	}
	if sent >= config.Max {
		return false
	}
	// Silence is measured from the last time we WROTE to them, never from the
	// record's `updated`: regenerating a checkout link saves the record, so using
	// `updated` made every reminder reset its own clock and nobody was ever due.
	last := registration.GetDateTime("created").Time()
	if raw, ok := data["payment_reminder_at"].(string); ok && strings.TrimSpace(raw) != "" {
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			last = parsed
		}
	}
	if time.Since(last) < time.Duration(config.AfterDays)*24*time.Hour {
		log.Printf("[retreats] %s: not due yet (%d reminders sent, last contact %s)", registration.GetString("email"), sent, last.Format("2006-01-02 15:04"))
		return false
	}

	// The attempt is written down BEFORE the email leaves: if the send fails the
	// person waits another round instead of being retried every few minutes, and
	// if the bookkeeping fails afterwards nobody gets a second copy today.
	if err := markReminderAttempt(app, registration, sent+1); err != nil {
		app.Logger().Warn("retreats: reminder bookkeeping failed", "error", err, "registration", registration.Id)
		return false
	}

	url, err := ResumeCheckout(app, retreat, registration)
	if err != nil {
		app.Logger().Warn("retreats: reminder checkout failed", "error", err, "registration", registration.Id)
		log.Printf("[retreats] %s: checkout failed: %v", registration.GetString("email"), err)
		return false
	}
	if !SendPaymentLinkEmail(app, retreat, registration.GetString("email"), url) {
		log.Printf("[retreats] %s: reminder email not sent (template or sender missing)", registration.GetString("email"))
		return false
	}

	log.Printf("[retreats] %s: reminder %d of %d sent", registration.GetString("email"), sent+1, config.Max)
	return true
}

// markReminderAttempt records that we are writing to this person now.
func markReminderAttempt(app *pocketbase.PocketBase, registration *core.Record, count int) error {
	data := backendinternal.ParseJSONMap(registration.Get("data"))
	data["payment_reminders_sent"] = count
	data["payment_reminder_at"] = time.Now().UTC().Format(time.RFC3339)
	registration.Set("data", data)
	return app.Save(registration)
}

// StartPaymentRemindersSchedule chases deposits once a day, in the same morning
// slot as the figures: the organiser's day starts with the email and the people
// who owe money hear from us at a civil hour, not at whatever time the container
// happened to restart.
func StartPaymentRemindersSchedule(app *pocketbase.PocketBase) {
	startMorningSchedule("checking deposit reminders", func() { SendPaymentReminders(app) })
}
