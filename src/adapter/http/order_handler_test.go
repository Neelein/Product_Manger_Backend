package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type orderServiceStub struct{ createdItems []model.OrderItem }

func (s *orderServiceStub) Create(_ context.Context, _ *model.Member, input usecase.OrderCreateInput) (*model.Order, error) {
	for _, i := range input.Items {
		s.createdItems = append(s.createdItems, model.OrderItem{ProductPriceID: i.ProductPriceID, Quantity: i.Quantity})
	}
	items := append([]model.OrderItem(nil), s.createdItems...)
	items[0].ProductSnapshot = []byte(`{"product_name":"Snapshot Product","image_url":"/media/images/products/product-1/image.jpg"}`)
	return &model.Order{ID: "order-1", OrderNo: "ORD-1", CustomerSnapshot: []byte(`{"name":"Checkout User","phone":"0912345678","email":"checkout@example.com"}`), ShippingAddressSnapshot: []byte(`{"delivery_method":"email"}`), Items: items}, nil
}
func (s *orderServiceStub) GetByID(context.Context, *model.Member, string) (*model.Order, error) {
	return &model.Order{}, nil
}
func (s *orderServiceStub) List(context.Context, *model.Member, string, int, int) ([]model.Order, int, error) {
	return nil, 0, nil
}
func (s *orderServiceStub) Cancel(context.Context, *model.Member, string) error { return nil }
func (s *orderServiceStub) UpdateStatus(context.Context, *model.Member, string, string) (*model.Order, error) {
	return &model.Order{}, nil
}
func (s *orderServiceStub) History(context.Context, *model.Member, string) ([]model.OrderStatusHistory, error) {
	return nil, nil
}

func TestOrderHandlerCreateIgnoresClientPriceFields(t *testing.T) {
	stub := &orderServiceStub{}
	h := NewOrderHandler(stub)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(`{"items":[{"product_price_id":"price-1","quantity":2,"unit_price":0.01,"line_total":0.02}],"customer":{"name":"Checkout User","phone":"0912345678","email":"checkout@example.com"},"delivery_method":"email"}`))
	req = req.WithContext(ContextWithMember(req.Context(), &Member{ID: "customer-1", MemberType: "customer"}))
	rr := httptest.NewRecorder()
	h.Create(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code)
	require.Equal(t, "price-1", stub.createdItems[0].ProductPriceID)
	var response OrderResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	require.Equal(t, "ORD-1", response.Order.OrderNo)
	var customer map[string]string
	require.NoError(t, json.Unmarshal(response.Order.CustomerSnapshot, &customer))
	require.Equal(t, "checkout@example.com", customer["email"])
	var shipping map[string]string
	require.NoError(t, json.Unmarshal(response.Order.ShippingAddressSnapshot, &shipping))
	require.Equal(t, "email", shipping["delivery_method"])
	var product map[string]any
	require.NoError(t, json.Unmarshal(response.Order.Items[0].ProductSnapshot, &product))
	require.Equal(t, "Snapshot Product", product["product_name"])
	require.Equal(t, "/media/images/products/product-1/image.jpg", product["image_url"])
}

func TestOrderResponseSerializesMissingProductImageAsNull(t *testing.T) {
	response := orderResponseDTO(model.Order{Items: []model.OrderItem{{ProductSnapshot: []byte(`{"product_name":"No Image","image_url":null}`)}}})
	var snapshot map[string]any
	require.NoError(t, json.Unmarshal(response.Items[0].ProductSnapshot, &snapshot))
	require.Contains(t, snapshot, "image_url")
	require.Nil(t, snapshot["image_url"])
}

func TestOrderHandlerGetReadsRouteOrderID(t *testing.T) {
	stub := &orderServiceStub{}
	h := NewOrderHandler(stub)
	r := httptest.NewRequest(http.MethodGet, "/api/orders/order-1", nil)
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee-1", MemberType: "employee"}))
	r = mux.SetURLVars(r, map[string]string{"orderId": "order-1"})
	rr := httptest.NewRecorder()
	h.Get(rr, r)
	require.Equal(t, http.StatusOK, rr.Code)
}
