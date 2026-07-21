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
	for _, table := range []string{"read_receipts", "chat_messages", "chat_room_members", "chat_rooms", "announcements", "inventory_items", "inventories", "members", "product_prices", "product_details", "products"} {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS "+table+" CASCADE")
	}
	_, _ = pool.Exec(ctx, "DROP FUNCTION IF EXISTS create_chat_room, add_chat_room_members, get_chat_room_by_id, list_chat_rooms_by_member, update_chat_room, delete_chat_room, remove_chat_room_member, send_message, list_messages, delete_message, mark_message_read, get_message_read_by, count_room_unread, create_member, get_member_by_email, get_member_by_id, update_member, create_product, list_products, get_product_by_id, update_product, delete_product, create_product_detail, get_product_detail_by_product, update_product_detail, create_product_price, get_product_price_by_id, list_product_prices_by_detail, update_product_price, create_inventory, get_inventory_by_id, get_inventory_by_price_id, list_inventories, update_inventory, delete_inventory, create_inventory_item, get_inventory_item_by_id, list_inventory_items, update_inventory_item, delete_inventory_item, create_announcement, get_announcement_by_id, list_announcements, count_announcements, update_announcement, delete_announcement CASCADE")
}

func runMigration(ctx context.Context, pool *pgxpool.Pool) {
	for _, file := range []string{
		"../../../db/migrations/001_create_products.up.sql",
		"../../../db/migrations/002_create_members.up.sql",
		"../../../db/migrations/003_add_member_id_to_products.up.sql",
		"../../../db/migrations/004_create_inventory.up.sql",
		"../../../db/migrations/005_simplify_inventories.up.sql",
		"../../../db/migrations/006_create_functions.up.sql",
		"../../../db/migrations/007_add_inventory_id_to_price_functions.up.sql",
		"../../../db/migrations/008_create_announcements.up.sql",
		"../../../db/migrations/009_set_not_null.up.sql",
		"../../../db/migrations/010_create_chat.up.sql",
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
		Name:     "Test Product",
		Price:    29.99,
		Category: "electronics",
	}
	err := repo.Create(context.Background(), &p)
	require.NoError(t, err)
	return p
}

func TestProductRepositoryPGX_Create(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)

	product := domain.Product{
		Name:     "New Product",
		Price:    99.99,
		Category: "general",
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
		created.Category = "updated-category"

		err := repo.Update(context.Background(), &created)
		assert.NoError(t, err)

		updated, err := repo.GetByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
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

func cleanupDetails(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE product_details CASCADE")
	require.NoError(t, err)
}

func TestProductRepositoryPGX_CreateDetail(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)

	product := createTestProduct(t, repo)

	detail := domain.ProductDetail{
		ProductID:         product.ID,
		Introduction:      "這是一個優質商品",
		UsageInstructions: "使用前請詳閱說明",
		ReturnPolicy:      "七天鑑賞期",
	}

	err := repo.CreateDetail(context.Background(), &detail)
	assert.NoError(t, err)
	assert.NotEmpty(t, detail.ID)
	assert.Equal(t, product.ID, detail.ProductID)
	assert.Equal(t, "這是一個優質商品", detail.Introduction)
	assert.False(t, detail.CreatedAt.IsZero())
	assert.False(t, detail.UpdatedAt.IsZero())
}

func TestProductRepositoryPGX_CreatePrice(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)

	product := createTestProduct(t, repo)

	detail := domain.ProductDetail{
		ProductID: product.ID,
	}
	err := repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	price := domain.ProductPrice{
		ProductDetailID: detail.ID,
		Label:           "成人票",
		Amount:          100.00,
		Currency:        "TWD",
		SortOrder:       1,
	}

	err = repo.CreatePrice(context.Background(), &price)
	assert.NoError(t, err)
	assert.NotEmpty(t, price.ID)
	assert.Equal(t, detail.ID, price.ProductDetailID)
	assert.Equal(t, 100.00, price.Amount)
	assert.False(t, price.CreatedAt.IsZero())
	assert.False(t, price.UpdatedAt.IsZero())
}

func TestProductRepositoryPGX_GetDetailByProductID(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)

	product := createTestProduct(t, repo)

	t.Run("existing detail", func(t *testing.T) {
		created := domain.ProductDetail{
			ProductID:         product.ID,
			Introduction:      "介紹",
			UsageInstructions: "說明",
			ReturnPolicy:      "退貨",
		}
		err := repo.CreateDetail(context.Background(), &created)
		require.NoError(t, err)

		got, err := repo.GetDetailByProductID(context.Background(), product.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "介紹", got.Introduction)
	})

	t.Run("non-existent product", func(t *testing.T) {
		_, err := repo.GetDetailByProductID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrDetailNotFound)
	})
}

func TestProductRepositoryPGX_UpdateDetail(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)

	t.Run("update existing detail", func(t *testing.T) {
		created := domain.ProductDetail{
			ProductID: product.ID,
		}
		err := repo.CreateDetail(context.Background(), &created)
		require.NoError(t, err)
		originalUpdatedAt := created.UpdatedAt

		created.Introduction = "新介紹"
		created.UsageInstructions = "新說明"
		created.ReturnPolicy = "新退貨"

		err = repo.UpdateDetail(context.Background(), &created)
		assert.NoError(t, err)

		got, err := repo.GetDetailByProductID(context.Background(), product.ID)
		assert.NoError(t, err)
		assert.Equal(t, "新介紹", got.Introduction)
		assert.Equal(t, "新說明", got.UsageInstructions)
		assert.Equal(t, "新退貨", got.ReturnPolicy)
		assert.True(t, got.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("non-existent detail", func(t *testing.T) {
		d := domain.ProductDetail{
			ID: "00000000-0000-0000-0000-000000000000",
		}
		err := repo.UpdateDetail(context.Background(), &d)
		assert.ErrorIs(t, err, domain.ErrDetailNotFound)
	})
}

func TestProductRepositoryPGX_GetPriceByID(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)

	detail := domain.ProductDetail{ProductID: product.ID}
	err := repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	t.Run("existing price", func(t *testing.T) {
		created := domain.ProductPrice{
			ProductDetailID: detail.ID,
			Label:           "成人票",
			Amount:          100,
			Currency:        "TWD",
			SortOrder:       1,
		}
		err := repo.CreatePrice(context.Background(), &created)
		require.NoError(t, err)

		got, err := repo.GetPriceByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "成人票", got.Label)
		assert.Equal(t, 100.0, got.Amount)
	})

	t.Run("non-existent price", func(t *testing.T) {
		_, err := repo.GetPriceByID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrPriceNotFound)
	})
}

func TestProductRepositoryPGX_GetPricesByDetailID(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)

	detail := domain.ProductDetail{ProductID: product.ID}
	err := repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	t.Run("multiple prices ordered by sort_order", func(t *testing.T) {
		repo.CreatePrice(context.Background(), &domain.ProductPrice{
			ProductDetailID: detail.ID, Label: "B", Amount: 2, SortOrder: 2,
		})
		repo.CreatePrice(context.Background(), &domain.ProductPrice{
			ProductDetailID: detail.ID, Label: "A", Amount: 1, SortOrder: 1,
		})

		prices, err := repo.GetPricesByDetailID(context.Background(), detail.ID)
		assert.NoError(t, err)
		assert.Len(t, prices, 2)
		assert.Equal(t, 1, prices[0].SortOrder)
		assert.Equal(t, "A", prices[0].Label)
		assert.Equal(t, 2, prices[1].SortOrder)
		assert.Equal(t, "B", prices[1].Label)
	})

	t.Run("no prices", func(t *testing.T) {
		other := createTestProduct(t, repo)
		otherDetail := domain.ProductDetail{ProductID: other.ID}
		repo.CreateDetail(context.Background(), &otherDetail)

		prices, err := repo.GetPricesByDetailID(context.Background(), otherDetail.ID)
		assert.NoError(t, err)
		assert.Empty(t, prices)
	})
}

func TestProductRepositoryPGX_UpdatePrice(t *testing.T) {
	defer cleanupProducts(t)
	defer cleanupDetails(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)

	detail := domain.ProductDetail{ProductID: product.ID}
	err := repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	t.Run("update existing price", func(t *testing.T) {
		created := domain.ProductPrice{
			ProductDetailID: detail.ID,
			Label:           "原價",
			Amount:          100,
			Currency:        "TWD",
			SortOrder:       1,
		}
		err := repo.CreatePrice(context.Background(), &created)
		require.NoError(t, err)
		originalUpdatedAt := created.UpdatedAt

		created.Label = "特價"
		created.Amount = 80
		created.Currency = "USD"
		created.SortOrder = 2

		err = repo.UpdatePrice(context.Background(), &created)
		assert.NoError(t, err)

		got, err := repo.GetPriceByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "特價", got.Label)
		assert.Equal(t, 80.0, got.Amount)
		assert.Equal(t, "USD", got.Currency)
		assert.Equal(t, 2, got.SortOrder)
		assert.True(t, got.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("non-existent price", func(t *testing.T) {
		p := domain.ProductPrice{
			ID: "00000000-0000-0000-0000-000000000000",
		}
		err := repo.UpdatePrice(context.Background(), &p)
		assert.ErrorIs(t, err, domain.ErrPriceNotFound)
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
