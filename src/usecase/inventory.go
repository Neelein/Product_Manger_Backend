package usecase

import (
	"backend/src/domain/model"
	"context"
)

type InventoryVariantReader interface {
	GetVariantByID(context.Context, string) (*model.ProductVariant, error)
}

func ValidateInventoryVariant(ctx context.Context, reader InventoryVariantReader, variantID string) error {
	_, err := reader.GetVariantByID(ctx, variantID)
	return err
}
