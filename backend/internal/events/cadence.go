package events

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Cadence is stored on the event record as a small DSL string:
//
//	once                  → single occurrence at start_date
//	weekly:mon|tue|...    → every 7 days from start_date (start must already
//	                        fall on the named weekday — wizard ensures it)
//	monthly:1st-mon       → 1st Monday of each month, starting in start's month
//	monthly:2nd-sun       → 2nd Sunday of each month, …
//	monthly:last-fri      → last Friday of each month, …
//
// ComputeOccurrences expands a cadence + start + count into actual datetimes.
// Cancelled dates (from data.cancelled_dates) are NOT filtered here — that's
// a render-layer concern. This returns the canonical schedule.
const CadenceOnce = "once"

func ComputeOccurrences(start time.Time, cadence string, count int) ([]time.Time, error) {
	cadence = strings.TrimSpace(strings.ToLower(cadence))
	if cadence == "" || cadence == CadenceOnce {
		return []time.Time{start}, nil
	}
	if count < 1 {
		count = 1
	}

	if strings.HasPrefix(cadence, "weekly:") {
		weekday, err := parseWeekday(strings.TrimPrefix(cadence, "weekly:"))
		if err != nil {
			return nil, err
		}
		if start.Weekday() != weekday {
			return nil, fmt.Errorf("start_date weekday does not match cadence: %q", cadence)
		}
		out := make([]time.Time, count)
		for i := 0; i < count; i++ {
			out[i] = start.AddDate(0, 0, 7*i)
		}
		return out, nil
	}

	if strings.HasPrefix(cadence, "monthly:") {
		spec := strings.TrimPrefix(cadence, "monthly:")
		parts := strings.SplitN(spec, "-", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid monthly cadence: %q", cadence)
		}
		nthToken := strings.TrimSpace(parts[0])
		weekday, err := parseWeekday(parts[1])
		if err != nil {
			return nil, err
		}
		nth, isLast, err := parseNth(nthToken)
		if err != nil {
			return nil, err
		}
		firstDay := nthWeekdayOfMonth(start.Year(), start.Month(), weekday, nth, isLast)
		if start.Day() != firstDay {
			return nil, fmt.Errorf("start_date does not match cadence: %q", cadence)
		}
		out := make([]time.Time, 0, count)
		// Anchor month is the month of `start`. Each successive occurrence
		// advances by one calendar month.
		hh, mm, ss := start.Clock()
		for i := 0; i < count; i++ {
			month := start.AddDate(0, i, 0)
			day := nthWeekdayOfMonth(month.Year(), month.Month(), weekday, nth, isLast)
			t := time.Date(month.Year(), month.Month(), day, hh, mm, ss, 0, start.Location())
			out = append(out, t)
		}
		return out, nil
	}

	return nil, fmt.Errorf("unknown cadence: %q", cadence)
}

// LastOccurrence returns the date of the final occurrence (after count steps).
// Useful for past/future window filtering without expanding the full slice.
func LastOccurrence(start time.Time, cadence string, count int) (time.Time, error) {
	occs, err := ComputeOccurrences(start, cadence, count)
	if err != nil {
		return time.Time{}, err
	}
	if len(occs) == 0 {
		return start, nil
	}
	return occs[len(occs)-1], nil
}

// NextOccurrence returns the earliest occurrence at or after `now`, skipping
// any date present in `cancelled` (RFC3339 or YYYY-MM-DD strings). If every
// occurrence is in the past or cancelled, returns the last non-cancelled
// occurrence — letting the caller decide how to label it.
func NextOccurrence(start time.Time, cadence string, count int, now time.Time, cancelled []string) (time.Time, bool, error) {
	occs, err := ComputeOccurrences(start, cadence, count)
	if err != nil {
		return time.Time{}, false, err
	}
	skip := normalizeDates(cancelled)
	var lastFuture time.Time
	for _, t := range occs {
		if _, isCancelled := skip[t.Format("2006-01-02")]; isCancelled {
			continue
		}
		if !t.Before(now) {
			return t, true, nil
		}
		lastFuture = t
	}
	if lastFuture.IsZero() {
		return time.Time{}, false, nil
	}
	return lastFuture, false, nil
}

func normalizeDates(raw []string) map[string]struct{} {
	out := make(map[string]struct{}, len(raw))
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, r); err == nil {
			out[t.Format("2006-01-02")] = struct{}{}
			continue
		}
		if t, err := time.Parse("2006-01-02", r); err == nil {
			out[t.Format("2006-01-02")] = struct{}{}
			continue
		}
	}
	return out
}

func parseWeekday(token string) (time.Weekday, error) {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "mon":
		return time.Monday, nil
	case "tue":
		return time.Tuesday, nil
	case "wed":
		return time.Wednesday, nil
	case "thu":
		return time.Thursday, nil
	case "fri":
		return time.Friday, nil
	case "sat":
		return time.Saturday, nil
	case "sun":
		return time.Sunday, nil
	}
	return 0, fmt.Errorf("invalid weekday: %q", token)
}

func parseNth(token string) (int, bool, error) {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "last" {
		return 0, true, nil
	}
	// accept "1st", "2nd", "3rd", "4th", "5th"
	if len(token) >= 2 {
		nDigits := strings.TrimRightFunc(token, func(r rune) bool {
			return r == 's' || r == 't' || r == 'n' || r == 'r' || r == 'd' || r == 'h'
		})
		if n, err := strconv.Atoi(nDigits); err == nil && n >= 1 && n <= 5 {
			return n, false, nil
		}
	}
	return 0, false, fmt.Errorf("invalid nth marker: %q", token)
}

// nthWeekdayOfMonth returns the day-of-month for the Nth occurrence of
// `weekday` in `year/month`. If `isLast` is true, returns the LAST occurrence
// of that weekday in the month, ignoring `nth`.
func nthWeekdayOfMonth(year int, month time.Month, weekday time.Weekday, nth int, isLast bool) int {
	if isLast {
		// last day of month
		first := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
		last := first.Day()
		offset := (int(first.Weekday()) - int(weekday) + 7) % 7
		return last - offset
	}
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	offset := (int(weekday) - int(first.Weekday()) + 7) % 7
	day := 1 + offset + 7*(nth-1)
	// clamp: if month has fewer occurrences (e.g. asking for 5th Monday of a
	// month with only 4), fall back to the last available one.
	maxDays := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	for day > maxDays {
		day -= 7
	}
	return day
}
