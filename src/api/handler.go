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
	var req domain.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product := domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
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
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
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
