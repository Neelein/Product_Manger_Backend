package model

import "time"

type Inventory struct {
	ID               string    `json:"id"`
	ProductVariantID string    `json:"product_variant_id"`
	ProductPriceID   string    `json:"product_price_id"`
	ProductDetailID  string    `json:"product_detail_id"`
	ProductID        string    `json:"product_id"`
	Name             string    `json:"name"`
	VariantName      string    `json:"variant_name"`
	Status           string    `json:"status"`
	TotalQuantity    int       `json:"total_quantity"`
	SoldQuantity     int       `json:"sold_quantity"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
