package internal

import "testing"

// IsBetaTester is mostly a DB query; the only branches reachable without
// PocketBase fixtures are the two early returns: empty userID and nil app.
// Each must yield (false, nil) — never an error, never a true positive.
func TestIsBetaTesterEarlyReturns(t *testing.T) {
	cases := []struct {
		name   string
		userID string
	}{
		{"empty userID", ""},
		{"whitespace-only userID", "   "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := IsBetaTester(nil, tc.userID)
			if err != nil {
				t.Fatalf("IsBetaTester(nil, %q) returned error: %v", tc.userID, err)
			}
			if ok {
				t.Fatalf("IsBetaTester(nil, %q) = true, want false", tc.userID)
			}
		})
	}

	// Non-empty userID + nil app: still bails out via the `app == nil` guard.
	ok, err := IsBetaTester(nil, "user_abc")
	if err != nil {
		t.Fatalf("IsBetaTester(nil, %q) returned error: %v", "user_abc", err)
	}
	if ok {
		t.Fatalf("IsBetaTester(nil, %q) = true, want false (nil app guard)", "user_abc")
	}
}

// HasBattleplanAccess returns false for nil actor without ever calling the
// database — the only reachable branch without fixtures.
func TestHasBattleplanAccessNilActor(t *testing.T) {
	ok, err := HasBattleplanAccess(nil, nil)
	if err != nil {
		t.Fatalf("HasBattleplanAccess(nil, nil) returned error: %v", err)
	}
	if ok {
		t.Fatalf("HasBattleplanAccess(nil, nil) = true, want false")
	}
}
