package internal

import "testing"

func TestNormalizePhoneAcceptsMissingPlusPrefix(t *testing.T) {
	value, err := NormalizePhone("3930180219")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != "+39 30180219" {
		t.Fatalf("expected normalized phone, got %q", value)
	}
}

func TestNormalizePhoneAcceptsDoubleZeroPrefix(t *testing.T) {
	value, err := NormalizePhone("003930180219")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != "+39 30180219" {
		t.Fatalf("expected normalized phone, got %q", value)
	}
}

