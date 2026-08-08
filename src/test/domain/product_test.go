package domain_test

import (
	"testing"

	"backend/src/domain"

	"github.com/stretchr/testify/assert"
)

func TestProductStruct(t *testing.T) {
	product := domain.Product{
		ID:       "123",
		Name:     "Test Product",
		Price:    19.99,
		Category: "electronics",
	}

	assert.Equal(t, "123", product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, 19.99, product.Price)
	assert.Equal(t, "electronics", product.Category)
}

func TestCreateProductRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.CreateProductRequest
		wantName       string
		wantPrice      float64
		wantCategoryID string
	}{
		{
			name: "full request",
			request: domain.CreateProductRequest{
				Name:       "New Product",
				Price:      99.99,
				CategoryID: "cat-1",
			},
			wantName:       "New Product",
			wantPrice:      99.99,
			wantCategoryID: "cat-1",
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
			assert.Equal(t, tt.wantPrice, tt.request.Price)
			assert.Equal(t, tt.wantCategoryID, tt.request.CategoryID)
		})
	}
}

func TestUpdateProductRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.UpdateProductRequest
		wantName       string
		wantPrice      float64
		wantCategoryID string
	}{
		{
			name: "full update request",
			request: domain.UpdateProductRequest{
				Name:       "Updated Product",
				Price:      49.99,
				CategoryID: "cat-2",
			},
			wantName:       "Updated Product",
			wantPrice:      49.99,
			wantCategoryID: "cat-2",
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
			assert.Equal(t, tt.wantPrice, tt.request.Price)
			assert.Equal(t, tt.wantCategoryID, tt.request.CategoryID)
		})
	}
}
