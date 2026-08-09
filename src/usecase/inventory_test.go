package usecase

import (
	"backend/src/domain/model"
	"context"
	"testing"
)

type fakeInventoryVariants struct{ variant *model.ProductVariant }

func (f *fakeInventoryVariants) GetVariantByID(context.Context, string) (*model.ProductVariant, error) {
	return f.variant, nil
}

func TestValidateInventoryVariantUsesVariantRepository(t *testing.T) {
	if err := ValidateInventoryVariant(context.Background(), &fakeInventoryVariants{variant: &model.ProductVariant{ID: "v"}}, "v"); err != nil {
		t.Fatal(err)
	}
}
