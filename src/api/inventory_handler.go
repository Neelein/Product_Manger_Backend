package api

import (
	"context"
	"encoding/json"
	"net/http"

	"backend/src/domain"

	"github.com/gorilla/mux"
)

type InventoryHandler struct {
	repo domain.InventoryRepository
}

func NewInventoryHandler(repo domain.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

func (h *InventoryHandler) CreateInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inventory := domain.Inventory{
		ProductPriceID: req.ProductPriceID,
		Status:         req.Status,
	}

	if err := h.repo.CreateInventory(context.Background(), &inventory); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.InventoryResponse{Inventory: inventory})
}

func (h *InventoryHandler) ListInventories(w http.ResponseWriter, r *http.Request) {
	inventories, err := h.repo.ListInventories(context.Background())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryListResponse{Inventories: inventories})
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	inventory, err := h.repo.GetInventoryByID(context.Background(), inventoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryResponse{Inventory: *inventory})
}

func (h *InventoryHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventoryID := mux.Vars(r)["inventoryId"]

	inventory, err := h.repo.GetInventoryByID(context.Background(), inventoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req domain.UpdateInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inventory.Status = req.Status

	if err := h.repo.UpdateInventory(context.Background(), inventory); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryResponse{Inventory: *inventory})
}

func (h *InventoryHandler) DeleteInventory(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	inventoryID := mux.Vars(r)["inventoryId"]

	if err := h.repo.DeleteInventory(context.Background(), inventoryID); err != nil {
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

	var req domain.CreateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item := domain.InventoryItem{
		InventoryID: inventoryID,
		ItemCode:    req.ItemCode,
		Status:      req.Status,
		Cost:        req.Cost,
		DateAdded:   req.DateAdded,
	}

	if err := h.repo.CreateItem(context.Background(), &item); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.InventoryItemResponse{Item: item})
}

func (h *InventoryHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	inventoryID := mux.Vars(r)["inventoryId"]

	items, err := h.repo.ListItemsByInventoryID(context.Background(), inventoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryItemListResponse{Items: items})
}

func (h *InventoryHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemID := mux.Vars(r)["itemId"]

	item, err := h.repo.GetItemByID(context.Background(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryItemResponse{Item: *item})
}

func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := mux.Vars(r)["itemId"]

	item, err := h.repo.GetItemByID(context.Background(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var req domain.UpdateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item.ItemCode = req.ItemCode
	item.Status = req.Status
	item.Cost = req.Cost
	item.DateAdded = req.DateAdded

	if err := h.repo.UpdateItem(context.Background(), item); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.InventoryItemResponse{Item: *item})
}

func (h *InventoryHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID := mux.Vars(r)["itemId"]

	if err := h.repo.DeleteItem(context.Background(), itemID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "inventory item deleted"})
}
