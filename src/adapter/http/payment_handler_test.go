package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/domain/model"
	"backend/src/domain/repository"
	"backend/src/usecase"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type paymentServiceStub struct{}

type paymentRepositoryStub struct{ method string }

func (r *paymentRepositoryStub) Pay(_ context.Context, orderID, memberID, method, last4, masked string, now time.Time) (*model.Payment, error) {
	r.method = method
	return &model.Payment{OrderID: orderID, MemberID: memberID, Method: method, Status: "paid", Last4: last4, MaskedCard: masked, CreatedAt: now}, nil
}
func (*paymentRepositoryStub) ExpirePending(context.Context, time.Time, time.Time) (int, error) {
	return 0, nil
}

var _ repository.Payment = (*paymentRepositoryStub)(nil)

func (paymentServiceStub) Pay(_ context.Context, _ *model.Member, _ string, _ usecase.PaymentInput) (*model.Payment, error) {
	return &model.Payment{ID: "payment-1", Status: "paid", Last4: "4242"}, nil
}
func (paymentServiceStub) Expire(context.Context, time.Time) (int, error) { return 0, nil }

func TestPaymentHandlerReturnsPaymentWithoutCardSecrets(t *testing.T) {
	h := NewPaymentHandler(paymentServiceStub{})
	body, _ := json.Marshal(CreatePaymentRequest{Method: "credit_card", CardNumber: "4242424242424242", CVV: "123", OTP: "1234567"})
	req := httptest.NewRequest("POST", "/api/orders/order-1/payments", bytes.NewReader(body))
	req = req.WithContext(ContextWithMember(req.Context(), &Member{ID: "member-1"}))
	req = mux.SetURLVars(req, map[string]string{"orderId": "order-1"})
	w := httptest.NewRecorder()
	h.Pay(w, req)
	require.Equal(t, 201, w.Code)
	require.NotContains(t, w.Body.String(), "4242424242424242")
}

func TestPaymentHandlerAcceptsUppercaseMethodAndPersistsNormalizedValue(t *testing.T) {
	repo := &paymentRepositoryStub{}
	h := NewPaymentHandler(usecase.NewPaymentService(repo))
	body, _ := json.Marshal(CreatePaymentRequest{Method: "CREDIT_CARD", CardNumber: "4242424242424242", CVV: "123", OTP: "1234567"})
	req := httptest.NewRequest("POST", "/api/orders/order-1/payments", bytes.NewReader(body))
	req = req.WithContext(ContextWithMember(req.Context(), &Member{ID: "member-1"}))
	req = mux.SetURLVars(req, map[string]string{"orderId": "order-1"})
	w := httptest.NewRecorder()
	h.Pay(w, req)

	require.Equal(t, 201, w.Code)
	require.Equal(t, "credit_card", repo.method)
}
