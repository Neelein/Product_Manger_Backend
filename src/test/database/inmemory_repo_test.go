package database_test

import (
	"context"
	"testing"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryRepository_Create(t *testing.T) {
	repo := database.NewInMemoryRepository()

	product := domain.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       19.99,
	}

	err := repo.Create(context.Background(), &product)
	assert.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())
}

func TestInMemoryRepository_GetByID(t *testing.T) {
	repo := database.NewInMemoryRepository()

	created := domain.Product{
		Name:  "Test",
		Price: 10.0,
	}
	repo.Create(context.Background(), &created)

	tests := []struct {
		name      string
		id        string
		wantFound bool
	}{
		{
			name:      "existing product",
			id:        created.ID,
			wantFound: true,
		},
		{
			name:      "non-existent product",
			id:        "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tt.id)
			if tt.wantFound {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, got.ID)
			} else {
				assert.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestInMemoryRepository_List(t *testing.T) {
	t.Run("empty repository", func(t *testing.T) {
		repo := database.NewInMemoryRepository()

		products, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, products)
	})

	t.Run("repository with products", func(t *testing.T) {
		repo := database.NewInMemoryRepository()

		repo.Create(context.Background(), &domain.Product{Name: "A", Price: 10})
		repo.Create(context.Background(), &domain.Product{Name: "B", Price: 20})

		products, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, products, 2)
	})
}

func TestInMemoryRepository_Update(t *testing.T) {
	repo := database.NewInMemoryRepository()

	created := domain.Product{
		Name:  "Original",
		Price: 10.0,
	}
	repo.Create(context.Background(), &created)

	tests := []struct {
		name      string
		product   domain.Product
		wantFound bool
	}{
		{
			name: "update existing product",
			product: domain.Product{
				ID:    created.ID,
				Name:  "Updated",
				Price: 25.0,
			},
			wantFound: true,
		},
		{
			name: "update non-existent product",
			product: domain.Product{
				ID:    "nonexistent",
				Name:  "Nope",
				Price: 0,
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(context.Background(), &tt.product)
			if tt.wantFound {
				assert.NoError(t, err)
				assert.Equal(t, "Updated", tt.product.Name)
				assert.False(t, tt.product.UpdatedAt.IsZero())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestInMemoryRepository_Delete(t *testing.T) {
	repo := database.NewInMemoryRepository()

	created := domain.Product{
		Name:  "To Delete",
		Price: 5.0,
	}
	repo.Create(context.Background(), &created)

	t.Run("delete existing product", func(t *testing.T) {
		err := repo.Delete(context.Background(), created.ID)
		assert.NoError(t, err)
	})

	t.Run("delete non-existent product", func(t *testing.T) {
		err := repo.Delete(context.Background(), created.ID)
		assert.Error(t, err)
	})
}
