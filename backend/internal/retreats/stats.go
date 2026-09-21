package retreats

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	backendinternal "members/backend/internal"
	eventinternal "members/backend/internal/events"
)

// dailyStatsHour is when the daily figures go out, local time on the server.
const dailyStatsHour = 7

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

	// Reserved is what actually costs a seat: only the confirmed. Owing the deposit
	// or waiting for a call back are both waiting, and waiting takes no seat.
	Reserved int

	// The same people, by name and phone, so the organiser can act on the figure
	// instead of going to look up who is behind it.
	Confirmed []Person
	Awaiting  []Person
	Requests  []Person
}

// Person is one registrant as the organiser needs them: who they are, how to
// call them, and whether they are already one of the members.
type Person struct {
	Name   string
	Phone  string
	Email  string
	Member bool
	// Retries is how many times we went back to them after the first email.
	Retries int

	// Who they are, for the call: someone ringing a stranger wants to know the age
	// and the town before they dial, not after. The age, not the year they were
	// born — nobody wants to do the subtraction while the phone rings. Zero
	// means unknown.
	AgeYears      int
	Provenance    string
	MaritalStatus string

	// AcceptURL is the review page for a request still waiting on the
	// organiser: the list of people to call ends each one with the link that
	// settles them, so the daily email is where the deciding gets done.
	AcceptURL string
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
		name, phone := registrantDetails(app, record)
		person := Person{
			Name:    name,
			Phone:   phone,
			Email:   strings.TrimSpace(record.GetString("email")),
			Member:  strings.TrimSpace(record.GetString("user")) != "",
			Retries: PaymentRetries(record),

			AgeYears:      ageFromBirthYear(registrationField(record, "birth_year")),
			Provenance:    registrationField(record, "provenance"),
			MaritalStatus: registrantMaritalStatus(app, record),
		}
		switch record.GetString("status") {
		case "active":
			stats.Active++
			stats.Confirmed = append(stats.Confirmed, person)
			// `user` is set whenever the registrant was recognised as a member,
			// so its absence is what makes someone an outsider here.
			if strings.TrimSpace(record.GetString("user")) != "" {
				stats.Members++
			} else {
				stats.Guests++
			}
		case "awaiting_payment":
			stats.AwaitingPayment++
			stats.Awaiting = append(stats.Awaiting, person)
		case "pending":
			stats.Pending++
			if token := strings.TrimSpace(record.GetString("accept_token")); token != "" {
				person.AcceptURL = AcceptPageURL(app, token)
			}
			stats.Requests = append(stats.Requests, person)
		}
	}

	remaining, limited, err := RemainingCapacity(app, retreat)
	if err != nil {
		return stats, err
	}
	stats.Reserved = stats.Active
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
		retreatPlaceholders(retreat), statsPlaceholders(stats, templateLabels(app, TemplateKindAdminDailyStats))...,
	))
}

// listLabels are the few words the code has to write itself when it turns
// people into lines: an empty bucket, a missing name, an age, a retry count,
// the link that approves. They default to English like every other copy
// source, and the template record overrides them in whatever language the
// email is written in (`data.labels`, same keys).
type listLabels struct {
	Nobody  string
	NoName  string
	Age     string // "{n}" is the number of years
	Retries string // "{n}" is the number of retries
	Confirm string
}

func templateLabels(app *pocketbase.PocketBase, kind string) listLabels {
	labels := listLabels{
		Nobody:  "_nobody_",
		NoName:  "(no name)",
		Age:     "{n} years old",
		Retries: "{n} retries",
		Confirm: "Confirm",
	}
	template, found, err := eventinternal.LoadTemplateDataByKind(app, "", kind)
	if err != nil || !found {
		return labels
	}
	pick := func(key string, target *string) {
		if value := strings.TrimSpace(template.Labels[key]); value != "" {
			*target = value
		}
	}
	pick("nobody", &labels.Nobody)
	pick("no_name", &labels.NoName)
	pick("age", &labels.Age)
	pick("retries", &labels.Retries)
	pick("confirm", &labels.Confirm)
	return labels
}

func withCount(label string, n int) string {
	return strings.ReplaceAll(label, "{n}", strconv.Itoa(n))
}

// statsPlaceholders exposes the figures to the template, so the organiser can
// reword the email without touching this file.
func statsPlaceholders(stats Stats, labels listLabels) []string {
	remaining := "—"
	capacity := "—"
	if stats.Limited {
		remaining = fmt.Sprintf("%d", stats.Remaining)
		capacity = fmt.Sprintf("%d", stats.Capacity)
	}
	return []string{
		"[active]", fmt.Sprintf("%d", stats.Active),
		"[reserved]", fmt.Sprintf("%d", stats.Reserved),
		"[confirmed_list]", personLines(stats.Confirmed, false, false, labels),
		"[awaiting_list]", personLines(stats.Awaiting, true, false, labels),
		"[requests_list]", personLines(stats.Requests, false, true, labels),
		"[members]", fmt.Sprintf("%d", stats.Members),
		"[guests]", fmt.Sprintf("%d", stats.Guests),
		"[awaiting_payment]", fmt.Sprintf("%d", stats.AwaitingPayment),
		"[pending]", fmt.Sprintf("%d", stats.Pending),
		"[remaining]", remaining,
		"[capacity]", capacity,
	}
}

// ageFromBirthYear turns a birth year into an age in years. A year that is
// not a year, or one that would give an impossible age, comes back as 0 and
// is left out rather than shown wrong.
func ageFromBirthYear(raw string) int {
	year, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	age := time.Now().Year() - year
	if age <= 0 || age > 120 {
		return 0
	}
	return age
}

// registrationField reads one value the registrant typed into the form.
func registrationField(record *core.Record, key string) string {
	data := backendinternal.ParseJSONMap(record.Get("data"))
	return strings.TrimSpace(backendinternal.AnyToString(data[key]))
}

// registrantMaritalStatus is asked on the guest form; a member answered it in
// their profile instead, under the same key, so it is read from there when
// the form has nothing.
func registrantMaritalStatus(app *pocketbase.PocketBase, record *core.Record) string {
	if value := registrationField(record, "marital_status"); value != "" {
		return value
	}
	userID := strings.TrimSpace(record.GetString("user"))
	if userID == "" {
		return ""
	}
	user, err := app.FindRecordById("users", userID)
	if err != nil || user == nil {
		return ""
	}
	data := backendinternal.ParseJSONMap(user.Get("data"))
	return strings.TrimSpace(backendinternal.AnyToString(data["marital_status"]))
}

// personLines renders one person per line, name and phone, as markdown list
// items. An empty bucket says so rather than leaving a hole in the email.
func personLines(people []Person, showRetries, showOrigin bool, labels listLabels) string {
	if len(people) == 0 {
		return labels.Nobody
	}
	lines := make([]string, 0, len(people))
	for _, p := range people {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = labels.NoName
		}
		phone := strings.TrimSpace(p.Phone)
		head := name
		if showRetries && p.Retries > 0 {
			head += " (" + withCount(labels.Retries, p.Retries) + ")"
		}
		// Name on its own line, the ways to reach them underneath: on a phone one
		// long line of name, number and address wraps into unreadable soup.
		// Only the ways to reach them that exist: a dash where a phone number
		// should be is not information, it is noise with a separator attached.
		var contacts []string
		if phone != "" {
			contacts = append(contacts, phone)
		}
		if email := strings.TrimSpace(p.Email); email != "" {
			contacts = append(contacts, email)
		}
		// A list item, so the theme can rule a line between one person and the next
		// instead of the names running together down the page. The second line is
		// indented to stay inside the same item.
		line := "- **" + head + "**"
		if showOrigin {
			var origin []string
			if p.AgeYears > 0 {
				origin = append(origin, withCount(labels.Age, p.AgeYears))
			}
			if p.Provenance != "" {
				origin = append(origin, p.Provenance)
			}
			if p.MaritalStatus != "" {
				origin = append(origin, p.MaritalStatus)
			}
			if len(origin) > 0 {
				line += "  \n  " + strings.Join(origin, " · ")
			}
		}
		// One way of reaching them per line: a number and an address side by side
		// wrap into each other on a phone and neither can be tapped cleanly.
		for _, contact := range contacts {
			line += "  \n  " + contact
		}
		// Last line of the item: the way to settle this person from the phone
		// the email is being read on. Only requests still waiting carry one.
		if p.AcceptURL != "" {
			line += "  \n  [" + labels.Confirm + "](" + p.AcceptURL + ")"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// StartDailyStatsSchedule sends the figures once a day for every retreat that
// is still open. Same shape as the Telegram backfill schedule: a coarse ticker
// that checks the clock, rather than a cron dependency for one daily job.
//
// The day already sent is remembered in memory only: a restart on the same day
// after the mail went out sends it twice, which is noise, not damage. Missing
// it entirely would be worse.
func StartDailyStatsSchedule(app *pocketbase.PocketBase) {
	startMorningSchedule("daily registration stats", func() { sendDailyStatsForOpenRetreats(app) })
}

// startMorningSchedule runs a job once a day, at dailyStatsHour local time.
//
// A coarse ticker that looks at the clock, rather than a cron dependency for a
// job that runs once a day. Starting after today's slot waits for tomorrow
// instead of firing a late one the moment the container comes up, and the day
// already done is remembered in memory only: a restart on the same day sends a
// second copy, which is noise, not damage — missing the day entirely would be
// worse.
func startMorningSchedule(name string, run func()) {
	lastRun := ""
	if time.Now().Hour() >= dailyStatsHour {
		lastRun = time.Now().Format("2006-01-02")
	}
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			today := now.Format("2006-01-02")
			if now.Hour() < dailyStatsHour || today == lastRun {
				continue
			}
			log.Printf("[retreats] %s", name)
			run()
			lastRun = today
		}
	}()
}

// SendDailyStatsNow sends the figures immediately for every open retreat, so the
// email can be looked at without waiting for tomorrow morning's slot.
func SendDailyStatsNow(app *pocketbase.PocketBase) { sendDailyStatsForOpenRetreats(app) }

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
