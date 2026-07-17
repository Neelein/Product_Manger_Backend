package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"backend/src/domain"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	repo domain.ProductRepository
}

func NewProductHandler(repo domain.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) CreateProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product := domain.Product{
		Name:      req.Name,
		Status:    req.Status,
		Price:     req.Price,
		Category:  req.Category,
		CreatedBy: member.ID,
	}

	if err := h.repo.Create(context.Background(), &product); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.ProductResponse{Product: product})
}

func (h *ProductHandler) ListProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	products, err := h.repo.List(context.Background())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.ProductListResponse{Products: products})
}

func (h *ProductHandler) GetProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	product, err := h.repo.GetByID(context.Background(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.ProductResponse{Product: *product})
}

func (h *ProductHandler) UpdateProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	var req domain.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product := domain.Product{
		ID:       id,
		Name:     req.Name,
		Status:   req.Status,
		Price:    req.Price,
		Category: req.Category,
	}

	if err := h.repo.Update(context.Background(), &product); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.ProductResponse{Product: product})
}

func (h *ProductHandler) DeleteProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	if err := h.repo.Delete(context.Background(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "product deleted"})
}

func (h *ProductHandler) CreateDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	productID := mux.Vars(r)["id"]

	var req domain.CreateDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail := domain.ProductDetail{
		ProductID:         productID,
		Introduction:      req.Introduction,
		UsageInstructions: req.UsageInstructions,
		ReturnPolicy:      req.ReturnPolicy,
	}

	if err := h.repo.CreateDetail(context.Background(), &detail); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.DetailResponse{Detail: detail})
}

func (h *ProductHandler) CreatePrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	detailID := mux.Vars(r)["did"]

	var req domain.CreatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price := domain.ProductPrice{
		ProductDetailID: detailID,
		Label:           req.Label,
		Amount:          req.Amount,
		Currency:        req.Currency,
		SortOrder:       req.SortOrder,
	}

	if err := h.repo.CreatePrice(context.Background(), &price); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.PriceResponse{Price: price})
}

func (h *ProductHandler) GetDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["id"]

	detail, err := h.repo.GetDetailByProductID(context.Background(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.DetailResponse{Detail: *detail})
}

func (h *ProductHandler) UpdateDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := mux.Vars(r)["id"]

	detail, err := h.repo.GetDetailByProductID(context.Background(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req domain.UpdateDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail.Introduction = req.Introduction
	detail.UsageInstructions = req.UsageInstructions
	detail.ReturnPolicy = req.ReturnPolicy

	if err := h.repo.UpdateDetail(context.Background(), detail); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.DetailResponse{Detail: *detail})
}

func (h *ProductHandler) ListPrices(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["id"]

	detail, err := h.repo.GetDetailByProductID(context.Background(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	prices, err := h.repo.GetPricesByDetailID(context.Background(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.PriceListResponse{Prices: prices})
}

func (h *ProductHandler) GetPrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	pid := mux.Vars(r)["pid"]

	price, err := h.repo.GetPriceByID(context.Background(), pid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.PriceResponse{Price: *price})
}

func (h *ProductHandler) UpdatePrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pid := mux.Vars(r)["pid"]

	price, err := h.repo.GetPriceByID(context.Background(), pid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req domain.UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price.Label = req.Label
	price.Amount = req.Amount
	price.Currency = req.Currency
	price.SortOrder = req.SortOrder

	if err := h.repo.UpdatePrice(context.Background(), price); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.PriceResponse{Price: *price})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, domain.ErrorResponse{Error: message})
}
