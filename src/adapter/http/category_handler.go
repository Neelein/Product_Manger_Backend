package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/src/usecase"

	"github.com/gorilla/mux"
)

type CategoryHandler struct {
	repo usecase.CategoryService
}

func NewCategoryHandler(service usecase.CategoryService) *CategoryHandler {
	return &CategoryHandler{repo: service}
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CategoryListResponse{Categories: categories})
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "category name is required")
		return
	}

	category, err := h.repo.Create(r.Context(), req.Name)
	if err != nil {
		if errors.Is(err, ErrCategoryNameExists) {
			writeError(w, http.StatusConflict, "category name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CategoryResponse{Category: *category})
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "category name is required")
		return
	}

	updated, err := h.repo.Update(r.Context(), id, req.Name)
	if err != nil {
		if errors.Is(err, ErrCategoryNameExists) {
			writeError(w, http.StatusConflict, "category name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !updated {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	category, found := h.findCategory(r, id)
	if !found {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	writeJSON(w, http.StatusOK, CategoryResponse{Category: *category})
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	deleted, err := h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrCategoryInUse) {
			writeError(w, http.StatusConflict, "category in use")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "category not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "category deleted"})
}

func (h *CategoryHandler) findCategory(r *http.Request, id string) (*Category, bool) {
	categories, err := h.repo.List(r.Context())
	if err != nil {
		return nil, false
	}

	for i := range categories {
		if categories[i].ID == id {
			return &categories[i], true
		}
	}

	return nil, false
}
