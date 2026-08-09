package usecase

import (
	"backend/src/domain/model"
	"backend/src/domain/repository"
	"context"
	"fmt"
)

type ProductUseCase struct {
	products repository.Product
	variants repository.ProductVariants
}

func NewProductUseCase(products repository.Product, variants repository.ProductVariants) *ProductUseCase {
	return &ProductUseCase{products: products, variants: variants}
}

func (u *ProductUseCase) CreateVariant(ctx context.Context, productID, priceID string, optionIDs []string) error {
	detail, err := u.variants.GetDetailByProductID(ctx, productID)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(optionIDs))
	for _, id := range optionIDs {
		if _, ok := seen[id]; ok {
			return model.ErrDuplicateProductVariant
		}
		seen[id] = struct{}{}
		o, err := u.variants.GetOptionByID(ctx, id)
		if err != nil {
			return err
		}
		if o.ProductDetailID != detail.ID {
			return fmt.Errorf("option %s does not belong to product detail", id)
		}
	}
	return u.variants.CreateVariant(ctx, &model.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: priceID, OptionIDs: optionIDs})
}
