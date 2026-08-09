package usecase

import (
	"backend/src/domain/entity"
	"backend/src/domain/model"
)

func ValidateCategoryName(name string) error {
	if err := entity.RequiredName(name); err != nil {
		return err
	}
	return nil
}

func ValidateProductVariantIDs(priceID string, optionIDs []string) error {
	if err := entity.UUID(priceID); err != nil {
		return model.ErrInvalidProductVariant
	}
	seen := make(map[string]struct{}, len(optionIDs))
	for _, id := range optionIDs {
		if err := entity.UUID(id); err != nil {
			return model.ErrInvalidProductVariant
		}
		if _, exists := seen[id]; exists {
			return model.ErrDuplicateProductVariant
		}
		seen[id] = struct{}{}
	}
	return nil
}

func ValidateInventoryVariantID(id string) error {
	if err := entity.UUID(id); err != nil {
		return err
	}
	return nil
}
