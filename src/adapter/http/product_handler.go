package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	repo usecase.ProductService
}

func variantErrorStatus(err error) int {
	if errors.Is(err, ErrProductOptionNotFound) || errors.Is(err, ErrProductVariantNotFound) || errors.Is(err, ErrDetailNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrDuplicateProductVariant) || errors.Is(err, ErrInvalidProductVariant) {
		return http.StatusBadRequest
	}
	if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "does not belong") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func validVariantInput(req CreateProductVariantRequest) bool {
	if uuid.Validate(req.ProductPriceID) != nil {
		return false
	}
	for _, optionID := range req.OptionIDs {
		if uuid.Validate(optionID) != nil {
			return false
		}
	}
	return true
}

func (h *ProductHandler) routeDetail(r *http.Request) (*ProductDetail, error) {
	detail, err := h.repo.GetDetailByProductID(r.Context(), mux.Vars(r)["productId"])
	if err != nil {
		return nil, err
	}
	if id := mux.Vars(r)["detailId"]; id != "" && id != detail.ID {
		return nil, ErrDetailNotFound
	}
	return detail, nil
}

func (h *ProductHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	var req CreateProductOptionRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	o := ProductOption{ProductDetailID: detail.ID, Name: req.Name, Value: req.Value}
	if err := h.repo.CreateOption(r.Context(), &o); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ProductOptionResponse{Option: o})
}

func (h *ProductHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	detail, err := h.routeDetail(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	options, err := h.repo.ListOptionsByDetailID(r.Context(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionListResponse{Options: options})
}
func (h *ProductHandler) GetOption(w http.ResponseWriter, r *http.Request) {
	o, err := h.repo.GetOptionByID(r.Context(), mux.Vars(r)["optionId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != o.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductOptionNotFound.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionResponse{Option: *o})
}
func (h *ProductHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	o, err := h.repo.GetOptionByID(r.Context(), mux.Vars(r)["optionId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != o.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductOptionNotFound.Error())
		return
	}
	var req UpdateProductOptionRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	o.Name, o.Value = req.Name, req.Value
	if err = h.repo.UpdateOption(r.Context(), o); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionResponse{Option: *o})
}
func (h *ProductHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	o, err := h.repo.GetOptionByID(r.Context(), mux.Vars(r)["optionId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != o.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductOptionNotFound.Error())
		return
	}
	if err := h.repo.DeleteOption(r.Context(), o.ID); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product option deleted"})
}

func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	var req CreateProductVariantRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validVariantInput(req) {
		writeError(w, http.StatusBadRequest, ErrInvalidProductVariant.Error())
		return
	}
	v := ProductVariant{ProductDetailID: detail.ID, ProductPriceID: req.ProductPriceID, SKU: req.SKU, Status: req.Status, OptionIDs: req.OptionIDs}
	if err = h.repo.CreateVariant(r.Context(), &v); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ProductVariantResponse{Variant: v})
}
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	detail, err := h.routeDetail(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	vs, err := h.repo.ListVariantsByDetailID(r.Context(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantListResponse{Variants: vs})
}
func (h *ProductHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	v, err := h.repo.GetVariantByID(r.Context(), mux.Vars(r)["variantId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != v.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductVariantNotFound.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantResponse{Variant: *v})
}
func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	v, err := h.repo.GetVariantByID(r.Context(), mux.Vars(r)["variantId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != v.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductVariantNotFound.Error())
		return
	}
	var req UpdateProductVariantRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validVariantInput(CreateProductVariantRequest{ProductPriceID: req.ProductPriceID, OptionIDs: req.OptionIDs}) {
		writeError(w, http.StatusBadRequest, ErrInvalidProductVariant.Error())
		return
	}
	v.ProductPriceID, v.SKU, v.Status, v.OptionIDs = req.ProductPriceID, req.SKU, req.Status, req.OptionIDs
	if err = h.repo.UpdateVariant(r.Context(), v); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantResponse{Variant: *v})
}
func (h *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	v, err := h.repo.GetVariantByID(r.Context(), mux.Vars(r)["variantId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	detail, err := h.routeDetail(r)
	if err != nil || detail.ID != v.ProductDetailID {
		writeError(w, http.StatusNotFound, ErrProductVariantNotFound.Error())
		return
	}
	if err := h.repo.DeleteVariant(r.Context(), v.ID); err != nil {
		writeError(w, variantErrorStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product variant deleted"})
}

func NewProductHandler(service usecase.ProductService) *ProductHandler {
	return &ProductHandler{repo: service}
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

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product := Product{
		Name:       req.Name,
		Status:     req.Status,
		Price:      req.Price,
		CategoryID: req.CategoryID,
		CreatedBy:  member.ID,
	}

	if err := h.repo.Create(r.Context(), &product); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ProductResponse{Product: product})
}

func (h *ProductHandler) ListProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	products, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ProductListResponse{Products: products})
}

func (h *ProductHandler) GetProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	product, err := h.repo.GetByID(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ProductResponse{Product: *product})
}

func (h *ProductHandler) UpdateProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product := Product{
		ID:         productID,
		Name:       req.Name,
		Status:     req.Status,
		Price:      req.Price,
		CategoryID: req.CategoryID,
	}

	if err := h.repo.Update(r.Context(), &product); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ProductResponse{Product: product})
}

func (h *ProductHandler) DeleteProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	if err := h.repo.Delete(r.Context(), productID); err != nil {
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

	productID := mux.Vars(r)["productId"]

	var req CreateDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail := ProductDetail{
		ProductID:         productID,
		Introduction:      req.Introduction,
		UsageInstructions: req.UsageInstructions,
		ReturnPolicy:      req.ReturnPolicy,
	}

	if err := h.repo.CreateDetail(r.Context(), &detail); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, DetailResponse{Detail: detail})
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

	detailID := mux.Vars(r)["detailId"]

	var req CreatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price := ProductPrice{
		ProductDetailID: detailID,
		Label:           req.Label,
		Amount:          req.Amount,
		Currency:        req.Currency,
		SortOrder:       req.SortOrder,
	}

	if err := h.repo.CreatePrice(r.Context(), &price); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, PriceResponse{Price: price})
}

func (h *ProductHandler) GetDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	detail, err := h.repo.GetDetailByProductID(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DetailResponse{Detail: *detail})
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

	productID := mux.Vars(r)["productId"]

	detail, err := h.repo.GetDetailByProductID(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req UpdateDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail.Introduction = req.Introduction
	detail.UsageInstructions = req.UsageInstructions
	detail.ReturnPolicy = req.ReturnPolicy

	if err := h.repo.UpdateDetail(r.Context(), detail); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DetailResponse{Detail: *detail})
}

func (h *ProductHandler) ListPrices(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	detail, err := h.repo.GetDetailByProductID(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	prices, err := h.repo.GetPricesByDetailID(r.Context(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PriceListResponse{Prices: prices})
}

func (h *ProductHandler) GetPrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	priceID := mux.Vars(r)["priceId"]

	price, err := h.repo.GetPriceByID(r.Context(), priceID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PriceResponse{Price: *price})
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

	priceID := mux.Vars(r)["priceId"]

	price, err := h.repo.GetPriceByID(r.Context(), priceID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price.Label = req.Label
	price.Amount = req.Amount
	price.Currency = req.Currency
	price.SortOrder = req.SortOrder

	if err := h.repo.UpdatePrice(r.Context(), price); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PriceResponse{Price: *price})
}
