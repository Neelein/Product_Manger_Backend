package usecase

import (
	"context"
	"strings"
	"time"
	"unicode"

	"backend/src/domain/model"
	"backend/src/domain/repository"
)

const PaymentExpiration = 10 * time.Minute

type PaymentInput struct {
	Method     string
	CardNumber string
	CVV        string
	OTP        string
}

type PaymentService interface {
	Pay(context.Context, *model.Member, string, PaymentInput) (*model.Payment, error)
	Expire(context.Context, time.Time) (int, error)
}

type paymentService struct{ repository repository.Payment }

func NewPaymentService(repo repository.Payment) PaymentService {
	return &paymentService{repository: repo}
}

func validDigits(value string, min, max int) bool {
	if len(value) < min || len(value) > max {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validLuhn(value string) bool {
	sum, alternate := 0, false
	for i := len(value) - 1; i >= 0; i-- {
		digit := int(value[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}
	return sum%10 == 0
}

func validatePayment(input PaymentInput) (string, string, error) {
	method := strings.ToLower(strings.TrimSpace(input.Method))
	if method != "credit_card" {
		return "", "", model.ErrInvalidPayment
	}
	if !validDigits(input.CardNumber, 13, 19) || !validLuhn(input.CardNumber) {
		return "", "", model.ErrInvalidCardNumber
	}
	if !validDigits(input.CVV, 3, 4) {
		return "", "", model.ErrInvalidPayment
	}
	if input.OTP != "1234567" {
		return "", "", model.ErrInvalidOTP
	}
	return method, input.CardNumber[len(input.CardNumber)-4:], nil
}

func maskCard(last4 string) string { return "**** **** **** " + last4 }

func (s *paymentService) Pay(ctx context.Context, member *model.Member, orderID string, input PaymentInput) (*model.Payment, error) {
	if member == nil {
		return nil, model.ErrForbidden
	}
	method, last4, err := validatePayment(input)
	if err != nil {
		return nil, err
	}
	return s.repository.Pay(ctx, orderID, member.ID, method, last4, maskCard(last4), time.Now().UTC())
}

func (s *paymentService) Expire(ctx context.Context, now time.Time) (int, error) {
	return s.repository.ExpirePending(ctx, now.Add(-PaymentExpiration), now)
}

type PaymentWorker struct {
	service  PaymentService
	clock    func() time.Time
	interval time.Duration
}

func NewPaymentWorker(service PaymentService, clock func() time.Time) *PaymentWorker {
	if clock == nil {
		clock = time.Now
	}
	return &PaymentWorker{service: service, clock: clock, interval: 5 * time.Minute}
}

func (w *PaymentWorker) RunOnce(ctx context.Context) (int, error) {
	return w.service.Expire(ctx, w.clock().UTC())
}

func (w *PaymentWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = w.RunOnce(ctx)
		}
	}
}
