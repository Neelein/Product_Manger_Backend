package entity

import (
	"errors"

	"backend/src/domain/model"
)

func ValidateOrderItem(item model.OrderItem) error {
	if item.Quantity <= 0 {
		return errors.New("order item quantity must be positive")
	}
	if item.UnitPrice == "" || item.LineTotal == "" {
		return errors.New("order item amounts are required")
	}
	if len(item.ProductSnapshot) == 0 {
		return errors.New("product snapshot is required")
	}
	return nil
}
