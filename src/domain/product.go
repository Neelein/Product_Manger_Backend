package domain

import "time"

type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Price     float64   `json:"price"`
	Category  string    `json:"category"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type UpdateProductRequest struct {
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type ProductResponse struct {
	Product Product `json:"product"`
}

type ProductListResponse struct {
	Products []Product `json:"products"`
}

type ErrorResponse struct {
	Error string `json:"error"`
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
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateDetailRequest struct {
	Introduction      string `json:"introduction"`
	UsageInstructions string `json:"usage_instructions"`
	ReturnPolicy      string `json:"return_policy"`
}

type CreatePriceRequest struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	SortOrder int     `json:"sort_order"`
}

type UpdateDetailRequest struct {
	Introduction      string `json:"introduction"`
	UsageInstructions string `json:"usage_instructions"`
	ReturnPolicy      string `json:"return_policy"`
}

type UpdatePriceRequest struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	SortOrder int     `json:"sort_order"`
}

type DetailResponse struct {
	Detail ProductDetail `json:"detail"`
}

type PriceResponse struct {
	Price ProductPrice `json:"price"`
}

type PriceListResponse struct {
	Prices []ProductPrice `json:"prices"`
}
