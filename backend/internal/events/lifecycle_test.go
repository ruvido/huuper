package events

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestIsValidType(t *testing.T) {
	cases := map[string]bool{
		"rally":   true,
		"call":    true,
		"meetup":  true,
		"":        false,
		"unknown": false,
		"Rally":   false,
	}
	for input, want := range cases {
		if got := IsValidType(input); got != want {
			t.Errorf("IsValidType(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestIsAssistantCreatableType(t *testing.T) {
	cases := map[string]bool{
		"rally":  false,
		"call":   true,
		"meetup": true,
		"":       false,
	}
	for input, want := range cases {
		if got := IsAssistantCreatableType(input); got != want {
			t.Errorf("IsAssistantCreatableType(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestRequiresGroup(t *testing.T) {
	if RequiresGroup(TypeMeetup) != true {
		t.Errorf("meetup must require group")
	}
	if RequiresGroup(TypeCall) != false {
		t.Errorf("call must not require group (admin can create national)")
	}
	if RequiresGroup(TypeRally) != false {
		t.Errorf("rally must not require group")
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

func TestParseDatesDeduplicates(t *testing.T) {
	dates, err := parseDates([]string{"2026-05-04", "2026-05-11", "2026-05-04"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dates) != 2 {
		t.Fatalf("expected 2 unique dates, got %d", len(dates))
	}
}

func TestValidateCreate(t *testing.T) {
	cfg := &Config{Recurrence: RecurrenceDef{MaxOccurrences: 10}}

	rally := TypeDef{Value: TypeRally, RequiresTitle: true, RequiresGroup: false}
	call := TypeDef{Value: TypeCall, RequiresTitle: false, RequiresGroup: false}
	meetup := TypeDef{Value: TypeMeetup, RequiresTitle: false, RequiresGroup: true}

	t.Run("rally without title fails", func(t *testing.T) {
		err := validateCreate(CreateInput{Type: TypeRally, Dates: []string{"2026-05-04"}}, rally, cfg)
		if err == nil {
			t.Errorf("expected error for missing title")
		}
	})
	t.Run("rally with title passes", func(t *testing.T) {
		err := validateCreate(CreateInput{Type: TypeRally, Title: "Annual gathering", Dates: []string{"2026-05-04"}}, rally, cfg)
		if err != nil {
			t.Errorf("unexpected: %v", err)
		}
	})
	t.Run("call without group passes", func(t *testing.T) {
		err := validateCreate(CreateInput{Type: TypeCall, Dates: []string{"2026-05-04"}}, call, cfg)
		if err != nil {
			t.Errorf("unexpected: %v", err)
		}
	})
	t.Run("meetup without group fails", func(t *testing.T) {
		err := validateCreate(CreateInput{Type: TypeMeetup, Dates: []string{"2026-05-04"}}, meetup, cfg)
		if err == nil {
			t.Errorf("expected error for missing group")
		}
	})
	t.Run("no dates fails", func(t *testing.T) {
		err := validateCreate(CreateInput{Type: TypeCall, Dates: []string{}}, call, cfg)
		if err == nil {
			t.Errorf("expected error for missing dates")
		}
	})
	t.Run("over max fails", func(t *testing.T) {
		dates := make([]string, 15)
		for i := range dates {
			dates[i] = "2026-05-04"
		}
		err := validateCreate(CreateInput{Type: TypeCall, Dates: dates}, call, cfg)
		if err == nil {
			t.Errorf("expected error for too many occurrences")
		}
	})
}

func TestGenerateSlugFormat(t *testing.T) {
	date := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	a := generateSlug(TypeCall, date)
	b := generateSlug(TypeCall, date)
	if a == b {
		t.Errorf("slugs must include random suffix to avoid collisions")
	}
	if len(a) < len("call-2026-05-04-12345678") {
		t.Errorf("slug too short: %q", a)
	}
}

func TestCollapseSeriesPicksUpcoming(t *testing.T) {
	collection := newEventTestCollection()
	r1 := newEventRecord(collection, "2026-05-04", "series-a")
	r2 := newEventRecord(collection, "2026-05-11", "series-a")
	r3 := newEventRecord(collection, "2026-05-18", "series-a")
	standalone := newEventRecord(collection, "2026-05-06", "")

	out := collapseSeries([]*core.Record{r2, r3, r1, standalone}, WindowFuture)
	if len(out) != 2 {
		t.Fatalf("expected 2 cards (1 series + 1 standalone), got %d", len(out))
	}

	var seriesPick *core.Record
	for _, r := range out {
		if r.GetString("series") == "series-a" {
			seriesPick = r
		}
	}
	if seriesPick == nil {
		t.Fatalf("series collapse missing")
	}
	if seriesPick.Id != r1.Id {
		t.Errorf("expected earliest occurrence (r1), got %s", seriesPick.Id)
	}
}

func TestCollapseSeriesPicksLatestForPast(t *testing.T) {
	collection := newEventTestCollection()
	r1 := newEventRecord(collection, "2025-05-04", "series-b")
	r2 := newEventRecord(collection, "2025-05-11", "series-b")

	out := collapseSeries([]*core.Record{r1, r2}, WindowPast)
	if len(out) != 1 {
		t.Fatalf("expected 1 collapsed record, got %d", len(out))
	}
	if out[0].Id != r2.Id {
		t.Errorf("expected latest (r2) for past window, got %s", out[0].Id)
	}
}

func newEventTestCollection() *core.Collection {
	collection := core.NewBaseCollection("events")
	collection.Fields.Add(
		&core.TextField{Name: "type"},
		&core.TextField{Name: "series"},
		&core.DateField{Name: "event_date"},
	)
	return collection
}

func newEventRecord(collection *core.Collection, date string, series string) *core.Record {
	r := core.NewRecord(collection)
	r.Id = "id-" + date + "-" + series
	r.Set("type", TypeCall)
	r.Set("series", series)
	parsed, _ := types.ParseDateTime(date)
	r.Set("event_date", parsed)
	return r
}
