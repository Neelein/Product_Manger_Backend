package domain

import "time"

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	Category Category `json:"category"`
}

type CategoryListResponse struct {
	Categories []Category `json:"categories"`
}