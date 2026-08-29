package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
)

type PaymentHandler struct{ service usecase.PaymentService }

func NewPaymentHandler(service usecase.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func paymentError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrForbidden), errors.Is(err, model.ErrOrderForbidden):
		status = http.StatusForbidden
	case errors.Is(err, model.ErrOrderNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrPaymentAlreadyExists), errors.Is(err, model.ErrPaymentNotAllowed):
		status = http.StatusConflict
	case errors.Is(err, model.ErrInvalidPayment), errors.Is(err, model.ErrInvalidCardNumber), errors.Is(err, model.ErrInvalidOTP):
		status = http.StatusBadRequest
	}
	writeError(w, status, err.Error())
}

func (h *PaymentHandler) Pay(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	payment, err := h.service.Pay(r.Context(), MemberFromContext(r.Context()), mux.Vars(r)["orderId"], usecase.PaymentInput{Method: req.Method, CardNumber: req.CardNumber, CVV: req.CVV, OTP: req.OTP})
	if err != nil {
		paymentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, PaymentResponse{Payment: *payment})
}
