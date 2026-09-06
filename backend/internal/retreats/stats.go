package retreats

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// dailyStatsHour is when the daily figures go out, local time on the server.
const dailyStatsHour = 8

// Stats is the state of a retreat's registrations in one glance.
//
// Members and Guests split the confirmed ones by how they got in, which is the
// number the organiser actually plans around: a member registered and paid on
// their own, a guest was called back and approved by hand. The two waiting
// buckets are kept apart because they need opposite things — a guest in
// Pending is waiting for a phone call from the organiser, someone in
// AwaitingPayment is waiting on nobody but themselves.
type Stats struct {
	Active          int
	Members         int
	Guests          int
	AwaitingPayment int
	Pending         int
	Capacity        int
	Remaining       int
	Limited         bool
}

// CountRegistrations tallies a retreat's registrations by status and kind.
func CountRegistrations(app *pocketbase.PocketBase, retreat *core.Record) (Stats, error) {
	stats := Stats{}
	if retreat == nil {
		return stats, fmt.Errorf("missing retreat")
	}

	records, err := app.FindRecordsByFilter(
		"retreat_registrations",
		"retreat = {:retreat}",
		"",
		0, 0,
		map[string]any{"retreat": retreat.Id},
	)
	if err != nil {
		return stats, err
	}

	for _, record := range records {
		switch record.GetString("status") {
		case "active":
			stats.Active++
			// `user` is set whenever the registrant was recognised as a member,
			// so its absence is what makes someone an outsider here.
			if strings.TrimSpace(record.GetString("user")) != "" {
				stats.Members++
			} else {
				stats.Guests++
			}
		case "awaiting_payment":
			stats.AwaitingPayment++
		case "pending":
			stats.Pending++
		}
	}

	remaining, limited, err := RemainingCapacity(app, retreat)
	if err != nil {
		return stats, err
	}
	stats.Capacity = retreat.GetInt("capacity")
	stats.Remaining = remaining
	stats.Limited = limited

	return stats, nil
}

// SendDailyStats emails the organiser the current figures for one retreat.
func SendDailyStats(app *pocketbase.PocketBase, retreat *core.Record) {
	stats, err := CountRegistrations(app, retreat)
	if err != nil {
		app.Logger().Warn("retreats: daily stats count failed", "error", err, "retreat", retreat.Id)
		return
	}
	sendAdminTemplateEmail(app, TemplateKindAdminDailyStats, append(
		retreatPlaceholders(retreat), statsPlaceholders(stats)...,
	))
}

// statsPlaceholders exposes the figures to the template, so the organiser can
// reword the email without touching this file.
func statsPlaceholders(stats Stats) []string {
	remaining := "—"
	capacity := "—"
	if stats.Limited {
		remaining = fmt.Sprintf("%d", stats.Remaining)
		capacity = fmt.Sprintf("%d", stats.Capacity)
	}
	return []string{
		"[active]", fmt.Sprintf("%d", stats.Active),
		"[members]", fmt.Sprintf("%d", stats.Members),
		"[guests]", fmt.Sprintf("%d", stats.Guests),
		"[awaiting_payment]", fmt.Sprintf("%d", stats.AwaitingPayment),
		"[pending]", fmt.Sprintf("%d", stats.Pending),
		"[remaining]", remaining,
		"[capacity]", capacity,
	}
}

// StartDailyStatsSchedule sends the figures once a day for every retreat that
// is still open. Same shape as the Telegram backfill schedule: a coarse ticker
// that checks the clock, rather than a cron dependency for one daily job.
//
// The day already sent is remembered in memory only: a restart on the same day
// after the mail went out sends it twice, which is noise, not damage. Missing
// it entirely would be worse.
func StartDailyStatsSchedule(app *pocketbase.PocketBase) {
	lastSent := ""
	if time.Now().Hour() >= dailyStatsHour {
		// Started after today's slot: wait for tomorrow instead of firing a
		// late one the moment the container comes up.
		lastSent = time.Now().Format("2006-01-02")
	}

	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			today := now.Format("2006-01-02")
			if now.Hour() < dailyStatsHour || today == lastSent {
				continue
			}
			log.Printf("[retreats] sending daily registration stats")
			sendDailyStatsForOpenRetreats(app)
			lastSent = today
		}
	}()
}

func sendDailyStatsForOpenRetreats(app *pocketbase.PocketBase) {
	records, err := app.FindRecordsByFilter("retreats", "active = true", "start_date", 0, 0)
	if err != nil {
		app.Logger().Warn("retreats: daily stats lookup failed", "error", err)
		return
	}
	for _, retreat := range records {
		SendDailyStats(app, retreat)
	}
}
