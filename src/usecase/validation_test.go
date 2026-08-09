package usecase

import (
	"testing"

	"backend/src/domain/model"
)

func TestValidateProductVariantIDs(t *testing.T) {
	if err := ValidateProductVariantIDs("bad", nil); err != model.ErrInvalidProductVariant {
		t.Fatalf("err = %v", err)
	}
	if err := ValidateProductVariantIDs("00000000-0000-0000-0000-000000000001", []string{
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000002",
	}); err != model.ErrDuplicateProductVariant {
		t.Fatalf("err = %v", err)
	}
}
