package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	service usecase.ProductService
}

func variantErrorStatus(err error) int {
	if errors.Is(err, ErrProductOptionNotFound) || errors.Is(err, ErrProductVariantNotFound) || errors.Is(err, ErrDetailNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrDuplicateSKU) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrDuplicateProductVariant) || errors.Is(err, ErrInvalidProductVariant) {
		return http.StatusBadRequest
	}
	if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "does not belong") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func variantErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrDuplicateSKU):
		return "此 SKU 已存在，請使用其他 SKU"
	case errors.Is(err, ErrDuplicateProductVariant):
		return "此產品變體組合已存在"
	case errors.Is(err, ErrInvalidProductVariant):
		return "產品變體資料無效"
	case errors.Is(err, ErrProductVariantNotFound):
		return "找不到產品變體"
	case errors.Is(err, ErrProductOptionNotFound):
		return "找不到產品規格"
	case errors.Is(err, ErrDetailNotFound):
		return "找不到產品詳細資訊"
	case strings.Contains(err.Error(), "duplicate"), strings.Contains(err.Error(), "does not belong"):
		return "產品變體資料無效"
	default:
		log.Printf("product variant operation failed: %v", err)
		return "操作失敗，請稍後再試"
	}
}

func (h *ProductHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(err))
		return
	}
	var req CreateProductOptionRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	o := ProductOption{ProductDetailID: detail.ID, Name: req.Name, Value: req.Value}
	if err := h.service.CreateOption(r.Context(), &o); err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusCreated, ProductOptionResponse{Option: o})
}

func (h *ProductHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(err))
		return
	}
	options, err := h.service.ListOptionsByDetailID(r.Context(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionListResponse{Options: options})
}
func (h *ProductHandler) GetOption(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(ErrProductOptionNotFound))
		return
	}
	o, err := h.service.GetOptionForDetail(r.Context(), detail.ID, mux.Vars(r)["optionId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionResponse{Option: *o})
}
func (h *ProductHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(ErrProductOptionNotFound))
		return
	}
	var req UpdateProductOptionRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	o, err := h.service.UpdateOptionApplication(r.Context(), detail.ID, mux.Vars(r)["optionId"], usecase.ProductOptionUpdateInput{Name: req.Name, Value: req.Value})
	if err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, ProductOptionResponse{Option: *o})
}
func (h *ProductHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, ErrProductOptionNotFound.Error())
		return
	}
	if err := h.service.DeleteOptionApplication(r.Context(), detail.ID, mux.Vars(r)["optionId"]); err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product option deleted"})
}

func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(err))
		return
	}
	var req CreateProductVariantRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	v := ProductVariant{ProductDetailID: detail.ID, ProductPriceID: req.ProductPriceID, SKU: req.SKU, Status: req.Status, OptionIDs: req.OptionIDs}
	if err = h.service.CreateVariant(r.Context(), &v); err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusCreated, ProductVariantResponse{Variant: v})
}
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(err))
		return
	}
	vs, err := h.service.ListVariantsByDetailID(r.Context(), detail.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantListResponse{Variants: vs})
}
func (h *ProductHandler) GetVariant(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(ErrProductVariantNotFound))
		return
	}
	v, err := h.service.GetVariantForDetail(r.Context(), detail.ID, mux.Vars(r)["variantId"])
	if err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantResponse{Variant: *v})
}
func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(ErrProductVariantNotFound))
		return
	}
	var req UpdateProductVariantRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	v, err := h.service.UpdateVariantApplication(r.Context(), detail.ID, mux.Vars(r)["variantId"], usecase.ProductVariantUpdateInput{ProductPriceID: req.ProductPriceID, SKU: req.SKU, Status: req.Status, OptionIDs: req.OptionIDs})
	if err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, ProductVariantResponse{Variant: *v})
}
func (h *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	detail, err := h.service.GetDetailForRoute(r.Context(), mux.Vars(r)["productId"], mux.Vars(r)["detailId"])
	if err != nil {
		writeError(w, http.StatusNotFound, variantErrorMessage(ErrProductVariantNotFound))
		return
	}
	if err := h.service.DeleteVariantApplication(r.Context(), detail.ID, mux.Vars(r)["variantId"]); err != nil {
		writeError(w, variantErrorStatus(err), variantErrorMessage(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "product variant deleted"})
}

func NewProductHandler(service usecase.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
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

	if err := h.service.Create(r.Context(), &product); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ProductResponse{Product: product})
}

func (h *ProductHandler) ListProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	products, err := h.service.Search(r.Context(), r.URL.Query().Get("keyword"), r.URL.Query().Get("category_id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ProductListResponse{Products: products})
}

func (h *ProductHandler) UploadImages(w http.ResponseWriter, r *http.Request) {
	if MemberFromContext(r.Context()) == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 30<<20+1<<20)
	if err := r.ParseMultipartForm(30 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}
	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "at least one image is required")
		return
	}
	if len(files) > 3 {
		writeError(w, http.StatusBadRequest, "a maximum of 3 images may be uploaded per request")
		return
	}
	productID := mux.Vars(r)["productId"]
	uploads := make([]usecase.UploadInput, 0, len(files))
	dir := filepath.Join(os.Getenv("MEDIA_ROOT"), "images/products", productID)
	for _, header := range files {
		if header.Size > 10<<20 {
			writeError(w, http.StatusBadRequest, "each image must be 10 MB or smaller")
			return
		}
		file, err := header.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid image")
			return
		}
		data, err := io.ReadAll(io.LimitReader(file, 10<<20+1))
		file.Close()
		if err != nil || len(data) > 10<<20 {
			writeError(w, http.StatusBadRequest, "invalid image")
			return
		}
		contentType := http.DetectContentType(data)
		ext := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}[contentType]
		if ext == "" {
			writeError(w, http.StatusBadRequest, "only JPEG, PNG, and WebP images are supported")
			return
		}
		name := uuid.NewString() + ext
		uploads = append(uploads, usecase.UploadInput{Directory: dir, Filename: name, Content: bytes.NewReader(data)})
	}
	images, err := h.service.UploadImages(r.Context(), productID, uploads)
	if err != nil {
		if strings.Contains(err.Error(), "maximum") || strings.Contains(err.Error(), "limit exceeded") {
			writeError(w, http.StatusBadRequest, "a product may have at most 3 images")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not save images")
		return
	}
	for i := range images {
		images[i].URL = "/media/images/products/" + productID + "/" + images[i].Filename
	}
	writeJSON(w, http.StatusCreated, ProductImageListResponse{Images: images})
}

func (h *ProductHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.service.ListImages(r.Context(), mux.Vars(r)["productId"])
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range images {
		images[i].URL = "/media/images/products/" + images[i].ProductID + "/" + images[i].Filename
	}
	writeJSON(w, http.StatusOK, ProductImageListResponse{Images: images})
}

func (h *ProductHandler) GetProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID := mux.Vars(r)["productId"]

	product, err := h.service.GetByID(r.Context(), productID)
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

	if err := h.service.Update(r.Context(), &product); err != nil {
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

	if err := h.service.Delete(r.Context(), productID); err != nil {
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

	if err := h.service.CreateDetail(r.Context(), &detail); err != nil {
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

	if err := h.service.CreatePrice(r.Context(), &price); err != nil {
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

	detail, err := h.service.GetDetailByProductID(r.Context(), productID)
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

	var req UpdateDetailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := h.service.UpdateDetailApplication(r.Context(), productID, usecase.ProductDetailUpdateInput{Introduction: req.Introduction, UsageInstructions: req.UsageInstructions, ReturnPolicy: req.ReturnPolicy})
	if err != nil {
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

	detail, err := h.service.GetDetailByProductID(r.Context(), productID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	prices, err := h.service.GetPricesByDetailID(r.Context(), detail.ID)
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

	price, err := h.service.GetPriceByID(r.Context(), priceID)
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

	var req UpdatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	price, err := h.service.UpdatePriceApplication(r.Context(), priceID, usecase.ProductPriceUpdateInput{Label: req.Label, Amount: req.Amount, Currency: req.Currency, SortOrder: req.SortOrder})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, PriceResponse{Price: *price})
}
