//go:build integration

package database_test

import (
	"context"
	"testing"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupInventory(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE inventory_items, inventories, product_prices, product_details, products CASCADE")
	require.NoError(t, err)
}

func createTestPrice(t *testing.T, repo *database.ProductRepositoryPGX) domain.ProductPrice {
	t.Helper()
	product := domain.Product{Name: "Test Product", Price: 100}
	err := repo.Create(context.Background(), &product)
	require.NoError(t, err)

	detail := domain.ProductDetail{ProductID: product.ID}
	err = repo.CreateDetail(context.Background(), &detail)
	require.NoError(t, err)

	price := domain.ProductPrice{
		ProductDetailID: detail.ID,
		Label:           "Test Price",
		Amount:          100,
		SortOrder:       1,
	}
	err = repo.CreatePrice(context.Background(), &price)
	require.NoError(t, err)

	return price
}

func TestInventoryRepositoryPGX_CreateInventory(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)

	inventory := domain.Inventory{
		ProductPriceID: price.ID,
		Name:           "Test Inventory",
		TotalQuantity:  100,
		Status:         "銷售中",
	}

	err := invRepo.CreateInventory(context.Background(), &inventory)
	assert.NoError(t, err)
	assert.NotEmpty(t, inventory.ID)
	assert.Equal(t, 0, inventory.SoldQuantity)
	assert.False(t, inventory.CreatedAt.IsZero())
	assert.False(t, inventory.UpdatedAt.IsZero())
}

func TestInventoryRepositoryPGX_GetInventoryByID(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
		Name:           "Test Inventory",
		TotalQuantity:  50,
	}
	err := invRepo.CreateInventory(context.Background(), &created)
	require.NoError(t, err)

	t.Run("existing inventory", func(t *testing.T) {
		got, err := invRepo.GetInventoryByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Test Inventory", got.Name)
		assert.Equal(t, 50, got.TotalQuantity)
		assert.Equal(t, "銷售中", got.Status)
	})

	t.Run("non-existent inventory", func(t *testing.T) {
		_, err := invRepo.GetInventoryByID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrInventoryNotFound)
	})
}

func TestInventoryRepositoryPGX_GetInventoryByPriceID(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
		Name:           "Test Inventory",
		TotalQuantity:  50,
	}
	err := invRepo.CreateInventory(context.Background(), &created)
	require.NoError(t, err)

	t.Run("existing by price id", func(t *testing.T) {
		got, err := invRepo.GetInventoryByPriceID(context.Background(), price.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("non-existent price id", func(t *testing.T) {
		_, err := invRepo.GetInventoryByPriceID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrInventoryNotFound)
	})
}

func TestInventoryRepositoryPGX_ListInventories(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	t.Run("empty list", func(t *testing.T) {
		inventories, err := invRepo.ListInventories(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, inventories)
	})

	t.Run("with inventories", func(t *testing.T) {
		price1 := createTestPrice(t, repo)
		price2 := createTestPrice(t, repo)

		inv1 := domain.Inventory{ProductPriceID: price1.ID, Name: "A"}
		inv2 := domain.Inventory{ProductPriceID: price2.ID, Name: "B"}
		require.NoError(t, invRepo.CreateInventory(context.Background(), &inv1))
		require.NoError(t, invRepo.CreateInventory(context.Background(), &inv2))

		inventories, err := invRepo.ListInventories(context.Background())
		assert.NoError(t, err)
		assert.Len(t, inventories, 2)
	})
}

func TestInventoryRepositoryPGX_UpdateInventory(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
		Name:           "Original",
		TotalQuantity:  100,
	}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &created))
	originalUpdatedAt := created.UpdatedAt

	t.Run("update existing", func(t *testing.T) {
		created.Name = "Updated"
		created.TotalQuantity = 200
		created.Status = "完售"

		err := invRepo.UpdateInventory(context.Background(), &created)
		assert.NoError(t, err)

		got, err := invRepo.GetInventoryByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", got.Name)
		assert.Equal(t, 200, got.TotalQuantity)
		assert.Equal(t, "完售", got.Status)
		assert.True(t, got.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("update non-existent", func(t *testing.T) {
		inv := domain.Inventory{
			ID:   "00000000-0000-0000-0000-000000000000",
			Name: "Ghost",
		}
		err := invRepo.UpdateInventory(context.Background(), &inv)
		assert.ErrorIs(t, err, domain.ErrInventoryNotFound)
	})
}

func TestInventoryRepositoryPGX_DeleteInventory(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)

	created := domain.Inventory{
		ProductPriceID: price.ID,
		Name:           "To Delete",
	}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &created))

	t.Run("delete existing", func(t *testing.T) {
		err := invRepo.DeleteInventory(context.Background(), created.ID)
		assert.NoError(t, err)

		_, err = invRepo.GetInventoryByID(context.Background(), created.ID)
		assert.ErrorIs(t, err, domain.ErrInventoryNotFound)
	})

	t.Run("delete non-existent", func(t *testing.T) {
		err := invRepo.DeleteInventory(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrInventoryNotFound)
	})
}

func TestInventoryRepositoryPGX_CreateItem(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)
	inventory := domain.Inventory{ProductPriceID: price.ID, Name: "Test"}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inventory))

	item := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "ITEM-001",
		Status:      "可用",
		Cost:        50.00,
		DateAdded:   "2026-07-18",
	}

	err := invRepo.CreateItem(context.Background(), &item)
	assert.NoError(t, err)
	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "可用", item.Status)
	assert.False(t, item.StatusUpdatedAt.IsZero())
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.UpdatedAt.IsZero())
}

func TestInventoryRepositoryPGX_GetItemByID(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)
	inventory := domain.Inventory{ProductPriceID: price.ID, Name: "Test"}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inventory))

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "ITEM-001",
		Status:      "可用",
		Cost:        50.00,
		DateAdded:   "2026-07-18",
	}
	require.NoError(t, invRepo.CreateItem(context.Background(), &created))

	t.Run("existing item", func(t *testing.T) {
		got, err := invRepo.GetItemByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "ITEM-001", got.ItemCode)
		assert.Equal(t, 50.00, got.Cost)
	})

	t.Run("non-existent item", func(t *testing.T) {
		_, err := invRepo.GetItemByID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrInventoryItemNotFound)
	})
}

func TestInventoryRepositoryPGX_ListItemsByInventoryID(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)
	inventory := domain.Inventory{ProductPriceID: price.ID, Name: "Test"}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inventory))

	t.Run("empty list", func(t *testing.T) {
		items, err := invRepo.ListItemsByInventoryID(context.Background(), inventory.ID)
		assert.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("with items", func(t *testing.T) {
		item1 := domain.InventoryItem{InventoryID: inventory.ID, ItemCode: "A"}
		item2 := domain.InventoryItem{InventoryID: inventory.ID, ItemCode: "B"}
		require.NoError(t, invRepo.CreateItem(context.Background(), &item1))
		require.NoError(t, invRepo.CreateItem(context.Background(), &item2))

		items, err := invRepo.ListItemsByInventoryID(context.Background(), inventory.ID)
		assert.NoError(t, err)
		assert.Len(t, items, 2)
	})
}

func TestInventoryRepositoryPGX_UpdateItem(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)
	inventory := domain.Inventory{ProductPriceID: price.ID, Name: "Test"}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inventory))

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "OLD-CODE",
		Status:      "可用",
		Cost:        30.00,
		DateAdded:   "2026-07-18",
	}
	require.NoError(t, invRepo.CreateItem(context.Background(), &created))

	t.Run("update existing item", func(t *testing.T) {
		created.ItemCode = "NEW-CODE"
		created.Status = "出售"
		created.Cost = 100.00

		err := invRepo.UpdateItem(context.Background(), &created)
		assert.NoError(t, err)

		got, err := invRepo.GetItemByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "NEW-CODE", got.ItemCode)
		assert.Equal(t, "出售", got.Status)
		assert.Equal(t, 100.00, got.Cost)
	})

	t.Run("update non-existent item", func(t *testing.T) {
		item := domain.InventoryItem{
			ID: "00000000-0000-0000-0000-000000000000",
		}
		err := invRepo.UpdateItem(context.Background(), &item)
		assert.ErrorIs(t, err, domain.ErrInventoryItemNotFound)
	})
}

func TestInventoryRepositoryPGX_DeleteItem(t *testing.T) {
	defer cleanupInventory(t)
	repo := database.NewProductRepositoryPGX(testPool)
	invRepo := database.NewInventoryRepositoryPGX(testPool)

	price := createTestPrice(t, repo)
	inventory := domain.Inventory{ProductPriceID: price.ID, Name: "Test"}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inventory))

	created := domain.InventoryItem{
		InventoryID: inventory.ID,
		ItemCode:    "TO-DELETE",
	}
	require.NoError(t, invRepo.CreateItem(context.Background(), &created))

	t.Run("delete existing item", func(t *testing.T) {
		err := invRepo.DeleteItem(context.Background(), created.ID)
		assert.NoError(t, err)

		_, err = invRepo.GetItemByID(context.Background(), created.ID)
		assert.ErrorIs(t, err, domain.ErrInventoryItemNotFound)
	})

	t.Run("delete non-existent item", func(t *testing.T) {
		err := invRepo.DeleteItem(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrInventoryItemNotFound)
	})
}
