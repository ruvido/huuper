package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	eventinternal "members/backend/internal/events"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type eventNext struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	EventDate     string `json:"event_date"`
	Registrations int    `json:"registrations"`
	Pending       int    `json:"pending"`
}

type countItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ageBucket struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

var ageBuckets = []struct {
	label    string
	min, max int
}{
	{"18-24", 0, 24},
	{"25-34", 25, 34},
	{"35-44", 35, 44},
	{"45-54", 45, 54},
	{"55+", 55, 200},
}

type userStats struct {
	Total      int         `json:"total"`
	Active     int         `json:"active"`
	NoTelegram int         `json:"noTelegram"`
	NotActive  int         `json:"notActive"`
	Angeli     int         `json:"angeli"`
	AvgAge     int         `json:"avgAge"`
	ByRegion   []countItem `json:"byRegion"`
	ByAge      []ageBucket `json:"byAge"`
	Marital    []countItem `json:"marital"`
	Work       []countItem `json:"work"`
	Sports     []countItem `json:"sports"`
	Interests  []countItem `json:"interests"`
	Skills     []countItem `json:"skills"`
}

var itMonths = [...]string{"gen", "feb", "mar", "apr", "mag", "giu", "lug", "ago", "set", "ott", "nov", "dic"}

func SummaryHandler(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		users, err := app.FindRecordsByFilter("users", "status = 'approved'", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_users", err)
		}

		stats, guardianDates := aggregateUserStats(users)

		groups, err := app.FindRecordsByFilter("groups", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_groups", err)
		}
		eventCount, err := app.FindRecordsByFilter("events", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_events_count", err)
		}

		allRequests, err := app.FindRecordsByFilter("requests", "", "", 0, 0)
		if err != nil {
			return apis.NewBadRequestError("failed_requests_all", err)
		}

		series := buildSeries(users, allRequests, guardianDates, stats)

		events, err := eventinternal.ListForUser(app, e.Auth, eventinternal.WindowFuture)
		if err != nil {
			return apis.NewBadRequestError("failed_events", err)
		}

		var next *eventNext
		if len(events) > 0 {
			event := events[0]
			eventID := event.ID
			registrations, err := app.FindRecordsByFilter("event_registrations", fmt.Sprintf("event = '%s'", eventID), "", 0, 0)
			if err != nil {
				return apis.NewBadRequestError("failed_registrations", err)
			}
			pending, err := app.FindRecordsByFilter("event_registrations", fmt.Sprintf("event = '%s' && status = 'pending'", eventID), "", 0, 0)
			if err != nil {
				return apis.NewBadRequestError("failed_pending", err)
			}

			next = &eventNext{
				ID:            eventID,
				Title:         event.Title,
				EventDate:     event.StartDate,
				Registrations: len(registrations),
				Pending:       len(pending),
			}
		}

		return e.JSON(http.StatusOK, map[string]any{
			"users":  stats,
			"groups": map[string]any{"total": len(groups)},
			"events": map[string]any{"total": len(eventCount), "next": next},
			"series": series,
		})
	}
}

func aggregateUserStats(users []*core.Record) (userStats, []time.Time) {
	stats := userStats{Total: len(users)}
	nowYear := time.Now().Year()
	regionCounts := map[string]int{}
	workCounts := map[string]int{}
	maritalCounts := map[string]int{}
	sportsCounts := map[string]int{}
	interestsCounts := map[string]int{}
	skillsCounts := map[string]int{}
	buckets := make([]int, len(ageBuckets))
	ageSum := 0
	ageCount := 0
	guardianDates := make([]time.Time, 0, len(users))

	for _, user := range users {
		if user.GetString("status") == "approved" {
			stats.Active++
		} else {
			stats.NotActive++
		}
		if isTelegramMissing(user.Get("telegram")) {
			stats.NoTelegram++
		}

		data := parseUserData(user.Get("data"))
		if data == nil {
			continue
		}
		if v := stringField(data["region"]); v != "" {
			regionCounts[v]++
		}
		if v := stringField(data["work"]); v != "" {
			workCounts[v]++
		}
		if v := stringField(data["marital_status"]); v != "" {
			maritalCounts[v]++
		}
		for _, v := range stringSliceField(data["sports"]) {
			sportsCounts[v]++
		}
		for _, v := range stringSliceField(data["interests"]) {
			interestsCounts[v]++
		}
		for _, v := range stringSliceField(data["skills"]) {
			skillsCounts[v]++
		}
		if by := stringField(data["birth_year"]); by != "" {
			if yr, err := strconv.Atoi(by); err == nil && yr > 1900 && yr <= nowYear {
				age := nowYear - yr
				ageSum += age
				ageCount++
				for i, b := range ageBuckets {
					if age >= b.min && age <= b.max {
						buckets[i]++
						break
					}
				}
			}
		}
		if g, ok := data["guardian"].(map[string]any); ok {
			stats.Angeli++
			if ts := stringField(g["assigned_at"]); ts != "" {
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					guardianDates = append(guardianDates, t)
				}
			}
		}
	}

	if ageCount > 0 {
		stats.AvgAge = ageSum / ageCount
	}
	stats.ByRegion = fillMissing(mapToSorted(regionCounts), stats.Total)
	stats.Work = fillMissing(mapToSorted(workCounts), stats.Total)
	stats.Marital = fillMissing(mapToSorted(maritalCounts), stats.Total)
	stats.Sports = topN(mapToSorted(sportsCounts), 6)
	stats.Interests = topN(mapToSorted(interestsCounts), 6)
	stats.Skills = topN(mapToSorted(skillsCounts), 6)
	stats.ByAge = make([]ageBucket, 0, len(ageBuckets)+1)
	for i, b := range ageBuckets {
		stats.ByAge = append(stats.ByAge, ageBucket{Name: b.label, Count: buckets[i]})
	}
	if missing := stats.Total - ageCount; missing > 0 {
		stats.ByAge = append(stats.ByAge, ageBucket{Name: "nd", Count: missing})
	}
	return stats, guardianDates
}

func fillMissing(items []countItem, total int) []countItem {
	sum := 0
	for _, it := range items {
		sum += it.Count
	}
	if missing := total - sum; missing > 0 {
		items = append(items, countItem{Name: "nd", Count: missing})
	}
	return items
}

func buildSeries(users, requests []*core.Record, guardianDates []time.Time, stats userStats) map[string]any {
	now := time.Now()
	const days = 30
	const points = 6
	step := time.Duration(days/(points-1)) * 24 * time.Hour
	ends := make([]time.Time, points)
	for i := 0; i < points; i++ {
		ends[i] = now.Add(-time.Duration(points-1-i) * step)
	}

	cumulativeUsers := cumulativeByTime(users, ends)
	activeRequests := activeRequestsByTime(requests, ends)
	cumulativeAngeli := cumulativeBy(guardianDates, ends)

	labels := make([]string, points)
	for i := 0; i < points; i++ {
		if i == points-1 {
			labels[i] = "oggi"
		} else {
			labels[i] = fmt.Sprintf("%d %s", ends[i].Day(), itMonths[int(ends[i].Month())-1])
		}
	}

	deltaUtenti := cumulativeUsers[points-1] - cumulativeUsers[0]
	deltaRichieste := activeRequests[points-1] - activeRequests[0]
	deltaAngeli := cumulativeAngeli[points-1] - cumulativeAngeli[0]

	return map[string]any{
		"labels": labels,
		"weekly": map[string]any{
			"utenti":    cumulativeUsers,
			"richieste": activeRequests,
			"angeli":    cumulativeAngeli,
		},
		"totals": map[string]any{
			"utenti":    stats.Total,
			"richieste": activeRequests[points-1],
			"angeli":    stats.Angeli,
		},
		"delta": map[string]any{
			"utenti":    deltaUtenti,
			"richieste": deltaRichieste,
			"angeli":    deltaAngeli,
		},
	}
}

func activeRequestsByTime(records []*core.Record, ends []time.Time) []int {
	out := make([]int, len(ends))
	for _, r := range records {
		createdAt := r.GetDateTime("created").Time()
		closedAt := requestClosedAt(r)
		for i, end := range ends {
			if createdAt.After(end) {
				continue
			}
			if !closedAt.IsZero() && !closedAt.After(end) {
				continue
			}
			out[i]++
		}
	}
	return out
}

func requestClosedAt(record *core.Record) time.Time {
	data := parseJSONMap(record.Get("data"))
	if data == nil {
		return time.Time{}
	}

	adminApprovedAt := parseRFC3339(strings.TrimSpace(stringField(data["admin_approved_at"])))
	if !record.GetBool("rejected") {
		return adminApprovedAt
	}

	rejectedAt := time.Time{}
	if rejectedData, ok := data["rejected"].(map[string]any); ok {
		rejectedAt = parseRFC3339(strings.TrimSpace(stringField(rejectedData["rejected_at"])))
	}
	if rejectedAt.IsZero() {
		return adminApprovedAt
	}
	if adminApprovedAt.IsZero() {
		return rejectedAt
	}
	if rejectedAt.Before(adminApprovedAt) {
		return rejectedAt
	}
	return adminApprovedAt
}

func parseRFC3339(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func cumulativeByTime(records []*core.Record, ends []time.Time) []int {
	out := make([]int, len(ends))
	for _, r := range records {
		t := r.GetDateTime("created").Time()
		for i, end := range ends {
			if !t.After(end) {
				out[i]++
			}
		}
	}
	return out
}

func cumulativeBy(times []time.Time, ends []time.Time) []int {
	out := make([]int, len(ends))
	for _, t := range times {
		for i, end := range ends {
			if !t.After(end) {
				out[i]++
			}
		}
	}
	return out
}

func parseUserData(raw any) map[string]any {
	return parseJSONMap(raw)
}

func parseJSONMap(raw any) map[string]any {
	if raw == nil {
		return nil
	}
	var bs []byte
	switch v := raw.(type) {
	case types.JSONRaw:
		bs = []byte(v)
	case []byte:
		bs = v
	case string:
		bs = []byte(v)
	default:
		return nil
	}
	trimmed := bytes.TrimSpace(bs)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(trimmed, &out); err != nil {
		return nil
	}
	return out
}

func stringField(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func stringSliceField(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			continue
		}
		for _, part := range strings.FieldsFunc(s, func(r rune) bool {
			return r == '\n' || r == ',' || r == ';' || r == '/'
		}) {
			p := strings.TrimSpace(part)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func mapToSorted(m map[string]int) []countItem {
	items := make([]countItem, 0, len(m))
	for k, v := range m {
		items = append(items, countItem{Name: k, Count: v})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Name < items[j].Name
	})
	return items
}

func topN(items []countItem, n int) []countItem {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func isTelegramMissing(raw any) bool {
	if raw == nil {
		return true
	}

	switch v := raw.(type) {
	case types.JSONRaw:
		return isNullJSON(string(bytes.TrimSpace(v)))
	case string:
		return isNullJSON(v)
	case []byte:
		return isNullJSON(string(bytes.TrimSpace(v)))
	default:
		return false
	}
}

func isNullJSON(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "null" {
		return true
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return false
	}
	return parsed == nil
}
