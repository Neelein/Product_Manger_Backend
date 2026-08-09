package entity

import "testing"

func TestRequiredNameRejectsWhitespace(t *testing.T) {
	if err := RequiredName(" \t"); err != ErrRequiredName {
		t.Fatalf("err = %v", err)
	}
}

func TestUUIDAndPageLimit(t *testing.T) {
	if err := UUID("not-a-uuid"); err != ErrInvalidUUID {
		t.Fatalf("UUID err = %v", err)
	}
	if err := PageLimit(0, 20); err != ErrInvalidPage {
		t.Fatalf("page err = %v", err)
	}
	if err := PageLimit(1, 101); err != ErrInvalidLimit {
		t.Fatalf("limit err = %v", err)
	}
}
