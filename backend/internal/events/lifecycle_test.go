package events

import (
	"testing"
	"time"
)

func TestConfigTypeDef(t *testing.T) {
	cfg := Config{Types: []TypeDef{
		{Value: "call", Creators: []string{"admin", "assistant"}, Required: map[string]bool{"url": true}},
	}}
	def, ok := cfg.TypeDef("call")
	if !ok {
		t.Fatalf("expected type def")
	}
	if !def.AllowsCreator("assistant") {
		t.Fatalf("expected assistant creator")
	}
	if !def.Requires("url") {
		t.Fatalf("expected url required")
	}
	if def.Requires("location") {
		t.Fatalf("expected location not required")
	}
}

func TestParseDate(t *testing.T) {
	t.Run("ISO date only", func(t *testing.T) {
		got, err := parseDate("2026-05-04")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Year() != 2026 || got.Month() != time.May || got.Day() != 4 {
			t.Errorf("got %v", got)
		}
	})
	t.Run("RFC3339", func(t *testing.T) {
		_, err := parseDate("2026-05-04T20:30:00Z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("datetime-local form", func(t *testing.T) {
		_, err := parseDate("2026-05-04T20:30")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty rejected", func(t *testing.T) {
		if _, err := parseDate("   "); err == nil {
			t.Errorf("expected error on empty")
		}
	})
	t.Run("garbage rejected", func(t *testing.T) {
		if _, err := parseDate("not-a-date"); err == nil {
			t.Errorf("expected error on garbage")
		}
	})
}

func TestNormalizeURL(t *testing.T) {
	t.Run("normalizes valid web url", func(t *testing.T) {
		got, err := normalizeURL(" HTTPS://Example.COM/path?q=1 ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "https://example.com/path?q=1" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("rejects single-label host", func(t *testing.T) {
		if _, err := normalizeURL("HTTPS://ASDSD"); err == nil {
			t.Fatalf("expected error")
		}
	})
	t.Run("accepts localhost", func(t *testing.T) {
		got, err := normalizeURL("http://localhost:3000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://localhost:3000" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestGenerateSlugFormat(t *testing.T) {
	date := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	a := generateSlug("call", date)
	b := generateSlug("call", date)
	if a == b {
		t.Errorf("slugs must include random suffix to avoid collisions")
	}
	if len(a) < len("call-2026-05-04-12345678") {
		t.Errorf("slug too short: %q", a)
	}
}

func TestComputeOccurrencesOnce(t *testing.T) {
	start := time.Date(2026, 5, 4, 19, 0, 0, 0, time.UTC)
	got, err := ComputeOccurrences(start, "once", 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 1 || !got[0].Equal(start) {
		t.Errorf("once should return [start], got %v", got)
	}
}

func TestComputeOccurrencesWeekly(t *testing.T) {
	start := time.Date(2026, 5, 4, 19, 0, 0, 0, time.UTC) // Monday
	got, err := ComputeOccurrences(start, "weekly:mon", 3)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	want := []time.Time{
		time.Date(2026, 5, 4, 19, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 11, 19, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 18, 19, 0, 0, 0, time.UTC),
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 occurrences, got %d", len(got))
	}
	for i, w := range want {
		if !got[i].Equal(w) {
			t.Errorf("occurrence %d: got %v want %v", i, got[i], w)
		}
	}
}

func TestComputeOccurrencesMonthlyNthWeekday(t *testing.T) {
	// 2nd Sunday: May 2026 → Sun 10, Jun → Sun 14, Jul → Sun 12
	start := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	got, err := ComputeOccurrences(start, "monthly:2nd-sun", 3)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	want := []time.Time{
		time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC),
	}
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d", len(got))
	}
	for i, w := range want {
		if !got[i].Equal(w) {
			t.Errorf("occurrence %d: got %v want %v", i, got[i], w)
		}
	}
}

func TestComputeOccurrencesMonthlyLast(t *testing.T) {
	// last Friday: May 2026 → Fri 29, Jun → Fri 26, Jul → Fri 31
	start := time.Date(2026, 5, 29, 18, 0, 0, 0, time.UTC)
	got, err := ComputeOccurrences(start, "monthly:last-fri", 3)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	want := []time.Time{
		time.Date(2026, 5, 29, 18, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 26, 18, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 31, 18, 0, 0, 0, time.UTC),
	}
	for i, w := range want {
		if !got[i].Equal(w) {
			t.Errorf("occurrence %d: got %v want %v", i, got[i], w)
		}
	}
}

func TestComputeOccurrencesInvalidCadence(t *testing.T) {
	start := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	if _, err := ComputeOccurrences(start, "weekly:funday", 3); err == nil {
		t.Errorf("expected error on bad weekday")
	}
	if _, err := ComputeOccurrences(start, "monthly:99th-mon", 3); err == nil {
		t.Errorf("expected error on bad nth marker")
	}
	if _, err := ComputeOccurrences(start, "yearly:jan", 3); err == nil {
		t.Errorf("expected error on unknown cadence")
	}
}

func TestNextOccurrenceSkipsCancelled(t *testing.T) {
	start := time.Date(2026, 5, 4, 19, 0, 0, 0, time.UTC) // Mon
	now := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)    // Tue, after first
	next, ok, err := NextOccurrence(start, "weekly:mon", 3, now, []string{"2026-05-11"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !ok {
		t.Fatalf("expected a future occurrence")
	}
	want := time.Date(2026, 5, 18, 19, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Errorf("got %v, want %v (May 11 cancelled, May 18 next)", next, want)
	}
}

func TestLastOccurrence(t *testing.T) {
	start := time.Date(2026, 5, 4, 19, 0, 0, 0, time.UTC)
	got, err := LastOccurrence(start, "weekly:mon", 3)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	want := time.Date(2026, 5, 18, 19, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
