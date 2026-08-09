package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type InventoryHandler struct {
	repo usecase.InventoryService
}

func NewInventoryHandler(service usecase.InventoryService) *InventoryHandler {
	return &InventoryHandler{repo: service}
}

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.ProductVariantID) == "" || uuid.Validate(req.ProductVariantID) != nil {
		writeError(w, http.StatusBadRequest, "product_variant_id is required")
		return
	}

	inventory := Inventory{
		ProductVariantID: req.ProductVariantID,
		ProductPriceID:   req.ProductPriceID,
		Status:           req.Status,
	}

	if err := h.repo.CreateInventory(r.Context(), &inventory); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, InventoryResponse{Inventory: inventory})
}

func (h *InventoryHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	inventories, err := h.repo.ListInventories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryListResponse{Inventories: inventories})
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	inventory, err := h.repo.GetInventoryByID(r.Context(), inventoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryResponse{Inventory: *inventory})
}

func (h *InventoryHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventoryID := mux.Vars(r)["inventoryId"]

	inventory, err := h.repo.GetInventoryByID(r.Context(), inventoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inventory.Status = req.Status

	if err := h.repo.UpdateInventory(r.Context(), inventory); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryResponse{Inventory: *inventory})
}

func (h *InventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventoryID := mux.Vars(r)["inventoryId"]

	if err := h.repo.DeleteInventory(r.Context(), inventoryID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "inventory deleted"})
}

func (h *InventoryHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventoryID := mux.Vars(r)["inventoryId"]

	var req CreateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item := InventoryItem{
		InventoryID: inventoryID,
		ItemCode:    req.ItemCode,
		Status:      req.Status,
		Cost:        req.Cost,
		DateAdded:   req.DateAdded,
	}

	if err := h.repo.CreateItem(r.Context(), &item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, InventoryItemResponse{Item: item})
}

func (h *InventoryHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	items, err := h.repo.ListItemsByInventoryID(r.Context(), inventoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryItemListResponse{Items: items})
}

func (h *InventoryHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemID := mux.Vars(r)["itemId"]

	item, err := h.repo.GetItemByID(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryItemResponse{Item: *item})
}

func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := mux.Vars(r)["itemId"]

	item, err := h.repo.GetItemByID(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req UpdateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.ItemCode = req.ItemCode
	item.Status = req.Status
	item.Cost = req.Cost
	item.DateAdded = req.DateAdded

	if err := h.repo.UpdateItem(r.Context(), item); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryItemResponse{Item: *item})
}

func (h *InventoryHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := mux.Vars(r)["itemId"]

	if err := h.repo.DeleteItem(r.Context(), itemID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "inventory item deleted"})
}
