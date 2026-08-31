package http

import (
	"encoding/json"
	"net/http"

	"backend/src/usecase"

	"github.com/gorilla/mux"
)

type InventoryHandler struct {
	service usecase.InventoryService
}

func NewInventoryHandler(service usecase.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateInventoryRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	inventory := Inventory{
		ProductVariantID: req.ProductVariantID,
		Status:           req.Status,
	}

	if err := h.service.CreateInventory(r.Context(), &inventory); err != nil {
		if err == usecase.ErrInvalidInventoryVariant {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, InventoryResponse{Inventory: inventory})
}

func (h *InventoryHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	inventories, err := h.service.ListInventories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryListResponse{Inventories: inventories})
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	inventory, err := h.service.GetInventoryByID(r.Context(), inventoryID)
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

	var req UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inventory, err := h.service.UpdateInventoryApplication(r.Context(), inventoryID, usecase.InventoryUpdateInput{Status: req.Status})
	if err != nil {
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

	if err := h.service.DeleteInventoryApplication(r.Context(), inventoryID); err != nil {
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

	if err := h.service.CreateItem(r.Context(), &item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, InventoryItemResponse{Item: item})
}

func (h *InventoryHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	items, err := h.service.ListItemsByInventoryID(r.Context(), inventoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, InventoryItemListResponse{Items: items})
}

func (h *InventoryHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemID := mux.Vars(r)["itemId"]

	item, err := h.service.GetItemByID(r.Context(), itemID)
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

	var req UpdateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.UpdateItemApplication(r.Context(), itemID, usecase.InventoryItemUpdateInput{ItemCode: req.ItemCode, Status: req.Status, Cost: req.Cost, DateAdded: req.DateAdded})
	if err != nil {
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

	if err := h.service.DeleteItemApplication(r.Context(), itemID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "inventory item deleted"})
}
