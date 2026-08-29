package model

import "time"

type Product struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Price      float64   `json:"price"`
	CategoryID string    `json:"category_id"`
	Category   string    `json:"category"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProductImage struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
}
type ProductDetail struct {
	ID                string    `json:"id"`
	ProductID         string    `json:"product_id"`
	Introduction      string    `json:"introduction"`
	UsageInstructions string    `json:"usage_instructions"`
	ReturnPolicy      string    `json:"return_policy"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type ProductPrice struct {
	ID              string    `json:"id"`
	ProductDetailID string    `json:"product_detail_id"`
	Label           string    `json:"label"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	SortOrder       int       `json:"sort_order"`
	InventoryID     *string   `json:"inventory_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type ProductOption struct {
	ID              string    `json:"id"`
	ProductDetailID string    `json:"product_detail_id"`
	Name            string    `json:"name"`
	Value           string    `json:"value"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type ProductVariant struct {
	ID              string    `json:"id"`
	ProductDetailID string    `json:"product_detail_id"`
	ProductPriceID  string    `json:"product_price_id"`
	SKU             *string   `json:"sku"`
	Status          string    `json:"status"`
	OptionIDs       []string  `json:"option_ids"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
