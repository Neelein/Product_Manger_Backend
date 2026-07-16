package domain_test

import (
	"testing"

	"backend/src/domain"

	"github.com/stretchr/testify/assert"
)

func TestProductStruct(t *testing.T) {
	product := domain.Product{
		ID:          "123",
		Name:        "Test Product",
		Description: "A test product",
		Price:       19.99,
		Category:    "electronics",
	}

	assert.Equal(t, "123", product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "A test product", product.Description)
	assert.Equal(t, 19.99, product.Price)
	assert.Equal(t, "electronics", product.Category)
}

func TestCreateProductRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     domain.CreateProductRequest
		wantName    string
		wantDesc    string
		wantPrice   float64
		wantCategory string
	}{
		{
			name: "full request",
			request: domain.CreateProductRequest{
				Name:        "New Product",
				Description: "Brand new",
				Price:       99.99,
				Category:    "books",
			},
			wantName:    "New Product",
			wantDesc:    "Brand new",
			wantPrice:   99.99,
			wantCategory: "books",
		},
		{
			name: "empty fields",
			request: domain.CreateProductRequest{
				Name: "",
			},
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.request.Name)
			assert.Equal(t, tt.wantDesc, tt.request.Description)
			assert.Equal(t, tt.wantPrice, tt.request.Price)
			assert.Equal(t, tt.wantCategory, tt.request.Category)
		})
	}
}

func TestUpdateProductRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     domain.UpdateProductRequest
		wantName    string
		wantDesc    string
		wantPrice   float64
		wantCategory string
	}{
		{
			name: "full update request",
			request: domain.UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated description",
				Price:       49.99,
				Category:    "software",
			},
			wantName:    "Updated Product",
			wantDesc:    "Updated description",
			wantPrice:   49.99,
			wantCategory: "software",
		},
		{
			name: "partial update",
			request: domain.UpdateProductRequest{
				Name: "Only Name",
			},
			wantName: "Only Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.request.Name)
			assert.Equal(t, tt.wantDesc, tt.request.Description)
			assert.Equal(t, tt.wantPrice, tt.request.Price)
			assert.Equal(t, tt.wantCategory, tt.request.Category)
		})
	}
}
