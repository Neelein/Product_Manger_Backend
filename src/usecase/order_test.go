package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"backend/src/domain/model"
	"github.com/stretchr/testify/require"
)

type orderRepoStub struct {
	created *model.Order
	items   []model.OrderItem
	err     error
}

func (s *orderRepoStub) Create(_ context.Context, o *model.Order, items []model.OrderItem) error {
	s.created = o
	s.items = items
	return s.err
}
func (s *orderRepoStub) GetByID(context.Context, string, string, bool) (*model.Order, error) {
	return nil, s.err
}
func (s *orderRepoStub) List(context.Context, string, string, int, int, bool) ([]model.Order, int, error) {
	return nil, 0, s.err
}
func (s *orderRepoStub) Cancel(context.Context, string, string, bool) error { return s.err }
func (s *orderRepoStub) UpdateStatus(context.Context, string, string, string) (*model.Order, error) {
	return nil, s.err
}
func (s *orderRepoStub) History(context.Context, string, string, bool) ([]model.OrderStatusHistory, error) {
	return nil, s.err
}

func TestOrderServiceCreateAllowsCustomerAndEmployeeAndDoesNotAcceptPrices(t *testing.T) {
	for _, member := range []*model.Member{{ID: "customer-1", MemberType: "customer"}, {ID: "employee-1", MemberType: "employee"}} {
		repo := &orderRepoStub{}
		svc := NewOrderService(repo)
		o, err := svc.Create(context.Background(), member, OrderCreateInput{Items: []OrderCreateItem{{ProductPriceID: "price-1", Quantity: 2}}, Customer: OrderCustomerInput{Name: "Name", Phone: "0912", Email: "name@example.com"}, DeliveryMethod: "email"})
		require.NoError(t, err)
		require.Equal(t, member.ID, o.CustomerID)
		require.Empty(t, repo.items[0].UnitPrice, "unit price must be resolved by the repository")
		require.Empty(t, repo.items[0].LineTotal, "line total must be resolved by the repository")
	}
}

func TestOrderServiceCreateRejectsInvalidItems(t *testing.T) {
	svc := NewOrderService(&orderRepoStub{})
	_, err := svc.Create(context.Background(), &model.Member{ID: "customer-1", MemberType: "customer"}, OrderCreateInput{Items: []OrderCreateItem{{ProductPriceID: "", Quantity: 1}}, Customer: OrderCustomerInput{Name: "Name", Phone: "0912", Email: "name@example.com"}, DeliveryMethod: "email"})
	require.ErrorIs(t, err, model.ErrInvalidOrder)
}

func TestOrderServiceCreateValidatesCheckoutDetails(t *testing.T) {
	svc := NewOrderService(&orderRepoStub{})
	base := OrderCreateInput{Items: []OrderCreateItem{{ProductPriceID: "price-1", Quantity: 1}}, Customer: OrderCustomerInput{Name: "Name", Phone: "0912", Email: "name@example.com"}, DeliveryMethod: "email"}
	tests := []struct {
		name   string
		mutate func(*OrderCreateInput)
	}{
		{"missing name", func(v *OrderCreateInput) { v.Customer.Name = " " }},
		{"missing phone", func(v *OrderCreateInput) { v.Customer.Phone = "" }},
		{"invalid email", func(v *OrderCreateInput) { v.Customer.Email = "not-an-email" }},
		{"invalid delivery method", func(v *OrderCreateInput) { v.DeliveryMethod = "pickup" }},
		{"missing home address", func(v *OrderCreateInput) { v.DeliveryMethod = "home_address" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base
			tt.mutate(&input)
			_, err := svc.Create(context.Background(), &model.Member{ID: "customer-1", MemberType: "customer"}, input)
			require.ErrorIs(t, err, model.ErrInvalidOrder)
		})
	}
}

func TestOrderServiceCreatePersistsSuppliedCheckoutSnapshots(t *testing.T) {
	repo := &orderRepoStub{}
	_, err := NewOrderService(repo).Create(context.Background(), &model.Member{ID: "customer-1", MemberType: "customer"}, OrderCreateInput{Items: []OrderCreateItem{{ProductPriceID: "price-1", Quantity: 1}}, Customer: OrderCustomerInput{Name: " Buyer ", Phone: " 0912 ", Email: "buyer@example.com"}, DeliveryMethod: "home_address", ShippingAddress: " Taipei "})
	require.NoError(t, err)
	var customer map[string]string
	require.NoError(t, json.Unmarshal(repo.created.CustomerSnapshot, &customer))
	require.Equal(t, map[string]string{"name": "Buyer", "phone": "0912", "email": "buyer@example.com"}, customer)
	var shipping map[string]string
	require.NoError(t, json.Unmarshal(repo.created.ShippingAddressSnapshot, &shipping))
	require.Equal(t, map[string]string{"delivery_method": "home_address", "address": "Taipei"}, shipping)
}

func TestOrderServiceEmployeeOnlyStatusTransition(t *testing.T) {
	svc := NewOrderService(&orderRepoStub{})
	_, err := svc.UpdateStatus(context.Background(), &model.Member{ID: "customer-1", MemberType: "customer"}, "order-1", "confirmed")
	require.ErrorIs(t, err, model.ErrForbidden)
}
