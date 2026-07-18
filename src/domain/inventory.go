package domain

import "time"

type Inventory struct {
	ID             string    `json:"id"`
	ProductPriceID string    `json:"product_price_id"`
	Name           string    `json:"name"`
	TotalQuantity  int       `json:"total_quantity"`
	SoldQuantity   int       `json:"sold_quantity"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InventoryItem struct {
	ID              string    `json:"id"`
	InventoryID     string    `json:"inventory_id"`
	ItemCode        string    `json:"item_code"`
	Status          string    `json:"status"`
	Cost            float64   `json:"cost"`
	DateAdded       string    `json:"date_added"`
	StatusUpdatedAt time.Time `json:"status_updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateInventoryRequest struct {
	ProductPriceID string `json:"product_price_id"`
	Name           string `json:"name"`
	TotalQuantity  int    `json:"total_quantity"`
	Status         string `json:"status"`
}

type UpdateInventoryRequest struct {
	Name          string `json:"name"`
	TotalQuantity int    `json:"total_quantity"`
	Status        string `json:"status"`
}

type CreateInventoryItemRequest struct {
	ItemCode  string  `json:"item_code"`
	Status    string  `json:"status"`
	Cost      float64 `json:"cost"`
	DateAdded string  `json:"date_added"`
}

type UpdateInventoryItemRequest struct {
	ItemCode  string  `json:"item_code"`
	Status    string  `json:"status"`
	Cost      float64 `json:"cost"`
	DateAdded string  `json:"date_added"`
}

type InventoryResponse struct {
	Inventory Inventory `json:"inventory"`
}

type InventoryListResponse struct {
	Inventories []Inventory `json:"inventories"`
}

type InventoryItemResponse struct {
	Item InventoryItem `json:"item"`
}

type InventoryItemListResponse struct {
	Items []InventoryItem `json:"items"`
}
