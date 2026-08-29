package usecase

import (
	"context"
	"testing"
	"time"

	"backend/src/domain/model"
	"github.com/stretchr/testify/require"
)

type paymentRepoStub struct {
	paid    int
	expired int
	now     time.Time
	method  string
}

func (r *paymentRepoStub) Pay(_ context.Context, orderID, memberID, method, last4, masked string, now time.Time) (*model.Payment, error) {
	r.paid++
	r.method = method
	return &model.Payment{OrderID: orderID, MemberID: memberID, Method: method, Status: "paid", Last4: last4, MaskedCard: masked, CreatedAt: now}, nil
}

func TestPaymentServiceNormalizesMethodBeforeRepository(t *testing.T) {
	repo := &paymentRepoStub{}
	service := NewPaymentService(repo)

	_, err := service.Pay(context.Background(), &model.Member{ID: "member-1"}, "order-1", PaymentInput{Method: "  CREDIT_CARD ", CardNumber: "4242424242424242", CVV: "123", OTP: "1234567"})

	require.NoError(t, err)
	require.Equal(t, "credit_card", repo.method)
}
func (r *paymentRepoStub) ExpirePending(_ context.Context, cutoff, now time.Time) (int, error) {
	r.expired++
	r.now = now
	_ = cutoff
	return 2, nil
}

func TestPaymentServiceValidatesFakeCardAndStoresOnlyLast4(t *testing.T) {
	repo := &paymentRepoStub{}
	service := NewPaymentService(repo)
	payment, err := service.Pay(context.Background(), &model.Member{ID: "member-1"}, "order-1", PaymentInput{Method: "credit_card", CardNumber: "4242424242424242", CVV: "123", OTP: "1234567"})
	require.NoError(t, err)
	require.Equal(t, 1, repo.paid)
	require.Equal(t, "4242", payment.Last4)
	require.Equal(t, "**** **** **** 4242", payment.MaskedCard)
}

func TestPaymentServiceRejectsInvalidCardAndOTP(t *testing.T) {
	service := NewPaymentService(&paymentRepoStub{})
	base := PaymentInput{Method: "credit_card", CardNumber: "4242424242424242", CVV: "123", OTP: "1234567"}
	_, err := service.Pay(context.Background(), &model.Member{ID: "m"}, "o", PaymentInput{Method: base.Method, CardNumber: "4242424242424243", CVV: base.CVV, OTP: base.OTP})
	require.ErrorIs(t, err, model.ErrInvalidCardNumber)
	base.OTP = "0000000"
	_, err = service.Pay(context.Background(), &model.Member{ID: "m"}, "o", base)
	require.ErrorIs(t, err, model.ErrInvalidOTP)
}

func TestPaymentWorkerRunOnceUsesInjectedClock(t *testing.T) {
	repo := &paymentRepoStub{}
	service := NewPaymentService(repo)
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	worker := NewPaymentWorker(service, func() time.Time { return now })
	count, err := worker.RunOnce(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.Equal(t, now, repo.now)
}
