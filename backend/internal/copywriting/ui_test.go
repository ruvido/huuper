package copywriting

import "testing"

func TestRequestButtonLabel(t *testing.T) {
	if got := RequestButtonLabel("  Accetta  ", "In verifica"); got != "Accetta" {
		t.Fatalf("expected cta to win, got %q", got)
	}
	if got := RequestButtonLabel("", "In verifica"); got != "In verifica" {
		t.Fatalf("expected label fallback, got %q", got)
	}
	if got := RequestButtonLabel("   ", "   "); got != DefaultRequestButtonLabel {
		t.Fatalf("expected default submit label, got %q", got)
	}
}
