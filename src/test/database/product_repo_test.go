//go:build integration

package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"backend/src/database"
	"backend/src/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://root:root123@localhost:5432/productdb_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	dropTables(ctx, pool)
	runMigration(ctx, pool)

	testPool = pool

	code := m.Run()

	pool.Close()
	os.Exit(code)
}

func dropTables(ctx context.Context, pool *pgxpool.Pool) {
	for _, table := range []string{"sessions", "members", "product_prices", "product_details", "products"} {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS "+table+" CASCADE")
	}
}

func runMigration(ctx context.Context, pool *pgxpool.Pool) {
	for _, file := range []string{
		"../../../db/migrations/001_create_products.sql",
		"../../../db/migrations/002_create_members.sql",
	} {
		schema, err := os.ReadFile(file)
		if err != nil {
			panic("failed to read migration file: " + err.Error())
		}
		_, err = pool.Exec(ctx, string(schema))
		if err != nil {
			panic("failed to run migration: " + err.Error())
		}
	}
}

func cleanupProducts(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE products CASCADE")
	require.NoError(t, err)
}

func createTestProduct(t *testing.T, repo *database.ProductRepositoryPGX) domain.Product {
	t.Helper()
	p := domain.Product{
		Name:        "Test Product",
		Description: "Test description",
		Price:       29.99,
		Category:    "electronics",
	}
	err := repo.Create(context.Background(), &p)
	require.NoError(t, err)
	return p
}

func TestProductRepositoryPGX_Create(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)

	product := domain.Product{
		Name:        "New Product",
		Description: "Brand new",
		Price:       99.99,
		Category:    "general",
	}

	err := repo.Create(context.Background(), &product)
	assert.NoError(t, err)
	assert.NotEmpty(t, product.ID)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())
}

func TestProductRepositoryPGX_GetByID(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	created := createTestProduct(t, repo)

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
			id:        "00000000-0000-0000-0000-000000000000",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tt.id)
			if tt.wantFound {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, got.ID)
				assert.Equal(t, created.Name, got.Name)
			} else {
				assert.ErrorIs(t, err, domain.ErrProductNotFound)
				assert.Nil(t, got)
			}
		})
	}
}

func TestProductRepositoryPGX_List(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)

	t.Run("empty repository", func(t *testing.T) {
		products, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, products)
	})

	t.Run("repository with products", func(t *testing.T) {
		p1 := createTestProduct(t, repo)
		p2 := createTestProduct(t, repo)

		products, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, products, 2)

		ids := make(map[string]bool, len(products))
		for _, p := range products {
			ids[p.ID] = true
		}
		assert.True(t, ids[p1.ID])
		assert.True(t, ids[p2.ID])
	})
}

func TestProductRepositoryPGX_Update(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)

	t.Run("update existing product", func(t *testing.T) {
		created := createTestProduct(t, repo)
		originalUpdatedAt := created.UpdatedAt

		created.Name = "Updated Name"
		created.Description = "Updated description"
		created.Category = "updated-category"

		err := repo.Update(context.Background(), &created)
		assert.NoError(t, err)

		updated, err := repo.GetByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, "updated-category", updated.Category)
		assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("update non-existent product", func(t *testing.T) {
		p := domain.Product{
			ID:   "00000000-0000-0000-0000-000000000000",
			Name: "Ghost",
		}
		err := repo.Update(context.Background(), &p)
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}

func TestProductRepositoryPGX_Delete(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)

	t.Run("delete existing product", func(t *testing.T) {
		created := createTestProduct(t, repo)
		err := repo.Delete(context.Background(), created.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(context.Background(), created.ID)
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})

	t.Run("delete non-existent product", func(t *testing.T) {
		err := repo.Delete(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}
