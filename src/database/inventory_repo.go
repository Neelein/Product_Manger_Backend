package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewInventoryRepositoryPGX(pool *pgxpool.Pool) *InventoryRepositoryPGX {
	return &InventoryRepositoryPGX{pool: pool}
}

func (r *InventoryRepositoryPGX) scanInventory(row pgx.Row) (*domain.Inventory, error) {
	var inv domain.Inventory
	err := row.Scan(
		&inv.ID, &inv.ProductPriceID, &inv.ProductDetailID, &inv.ProductID, &inv.Name, &inv.Status,
		&inv.TotalQuantity, &inv.SoldQuantity,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepositoryPGX) CreateInventory(
	ctx context.Context,
	inventory *domain.Inventory,
) error {
	status := inventory.Status
	if status == "" {
		status = "銷售中"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_inventory($1, $2)",
		inventory.ProductPriceID, status,
	).Scan(&inventory.ID, &inventory.CreatedAt, &inventory.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating inventory: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) GetInventoryByID(
	ctx context.Context,
	id string,
) (*domain.Inventory, error) {
	inv, err := r.scanInventory(r.pool.QueryRow(ctx, "SELECT * FROM get_inventory_by_id($1)", id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("getting inventory by ID: %w", err)
	}

	return inv, nil
}

func (r *InventoryRepositoryPGX) GetInventoryByPriceID(
	ctx context.Context,
	priceID string,
) (*domain.Inventory, error) {
	inv, err := r.scanInventory(r.pool.QueryRow(ctx, "SELECT * FROM get_inventory_by_price_id($1)", priceID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("getting inventory by price ID: %w", err)
	}

	return inv, nil
}

func (r *InventoryRepositoryPGX) ListInventories(
	ctx context.Context,
) ([]domain.Inventory, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_inventories()")
	if err != nil {
		return nil, fmt.Errorf("listing inventories: %w", err)
	}

	inventories, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Inventory, error) {
		var inv domain.Inventory
		err := row.Scan(
			&inv.ID, &inv.ProductPriceID, &inv.ProductDetailID, &inv.ProductID, &inv.Name, &inv.Status,
			&inv.TotalQuantity, &inv.SoldQuantity,
			&inv.CreatedAt, &inv.UpdatedAt,
		)
		return inv, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing inventories: %w", err)
	}

	if inventories == nil {
		inventories = []domain.Inventory{}
	}

	return inventories, nil
}

func (r *InventoryRepositoryPGX) UpdateInventory(
	ctx context.Context,
	inventory *domain.Inventory,
) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_inventory($1, $2)",
		inventory.ID, inventory.Status,
	).Scan(&inventory.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInventoryNotFound
		}
		return fmt.Errorf("updating inventory: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) DeleteInventory(
	ctx context.Context,
	id string,
) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_inventory($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting inventory: %w", err)
	}
	if !deleted {
		return domain.ErrInventoryNotFound
	}

	return nil
}

func (r *InventoryRepositoryPGX) CreateItem(
	ctx context.Context,
	item *domain.InventoryItem,
) error {
	dateAdded := item.DateAdded
	if dateAdded == "" {
		dateAdded = time.Now().Format("2006-01-02")
	}

	status := item.Status
	if status == "" {
		status = "可用"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_inventory_item($1, $2, $3, $4, $5)",
		item.InventoryID, item.ItemCode, status, item.Cost, dateAdded,
	).Scan(&item.ID, &item.StatusUpdatedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating inventory item: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) GetItemByID(
	ctx context.Context,
	id string,
) (*domain.InventoryItem, error) {
	var (
		item     domain.InventoryItem
		dateAddr time.Time
	)

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_inventory_item_by_id($1)", id,
	).Scan(&item.ID, &item.InventoryID, &item.ItemCode, &item.Status,
		&item.Cost, &dateAddr, &item.StatusUpdatedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryItemNotFound
		}
		return nil, fmt.Errorf("getting inventory item by ID: %w", err)
	}

	item.DateAdded = dateAddr.Format("2006-01-02")

	return &item, nil
}

func (r *InventoryRepositoryPGX) ListItemsByInventoryID(
	ctx context.Context,
	inventoryID string,
) ([]domain.InventoryItem, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_inventory_items($1)", inventoryID)
	if err != nil {
		return nil, fmt.Errorf("listing inventory items: %w", err)
	}

	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.InventoryItem, error) {
		var (
			item     domain.InventoryItem
			dateAddr time.Time
		)
		err := row.Scan(&item.ID, &item.InventoryID, &item.ItemCode, &item.Status,
			&item.Cost, &dateAddr, &item.StatusUpdatedAt,
			&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return item, err
		}
		item.DateAdded = dateAddr.Format("2006-01-02")
		return item, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing inventory items: %w", err)
	}

	if items == nil {
		items = []domain.InventoryItem{}
	}

	return items, nil
}

func (r *InventoryRepositoryPGX) UpdateItem(
	ctx context.Context,
	item *domain.InventoryItem,
) error {
	dateAdded := item.DateAdded
	if dateAdded == "" {
		dateAdded = time.Now().Format("2006-01-02")
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_inventory_item($1, $2, $3, $4, $5)",
		item.ID, item.ItemCode, item.Status, item.Cost, dateAdded,
	).Scan(&item.StatusUpdatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrInventoryItemNotFound
		}
		return fmt.Errorf("updating inventory item: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) DeleteItem(
	ctx context.Context,
	id string,
) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_inventory_item($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting inventory item: %w", err)
	}
	if !deleted {
		return domain.ErrInventoryItemNotFound
	}

	return nil
}
