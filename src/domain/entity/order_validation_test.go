package entity

import (
	"testing"

	"backend/src/domain/model"
)

func TestValidateOrderItem(t *testing.T) {
	valid := model.OrderItem{Quantity: 1, UnitPrice: "10.00", LineTotal: "10.00", ProductSnapshot: []byte(`{"id":"p1"}`)}
	if err := ValidateOrderItem(valid); err != nil {
		t.Fatal(err)
	}
	for name, item := range map[string]model.OrderItem{
		"zero quantity":    {Quantity: 0, UnitPrice: "1", LineTotal: "1", ProductSnapshot: []byte(`{}`)},
		"missing amount":   {Quantity: 1, LineTotal: "1", ProductSnapshot: []byte(`{}`)},
		"missing snapshot": {Quantity: 1, UnitPrice: "1", LineTotal: "1"},
	} {
		t.Run(name, func(t *testing.T) {
			if ValidateOrderItem(item) == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
