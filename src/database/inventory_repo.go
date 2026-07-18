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

func (r *InventoryRepositoryPGX) CreateInventory(
	ctx context.Context,
	inventory *domain.Inventory,
) error {
	status := inventory.Status
	if status == "" {
		status = "銷售中"
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO inventories (product_price_id, name, total_quantity, sold_quantity, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		inventory.ProductPriceID, inventory.Name, inventory.TotalQuantity,
		inventory.SoldQuantity, status,
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
	var inv domain.Inventory

	err := r.pool.QueryRow(ctx,
		`SELECT id, product_price_id, name, total_quantity, sold_quantity, status,
		        created_at, updated_at
		 FROM inventories WHERE id = $1`, id,
	).Scan(&inv.ID, &inv.ProductPriceID, &inv.Name, &inv.TotalQuantity,
		&inv.SoldQuantity, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("getting inventory by ID: %w", err)
	}

	return &inv, nil
}

func (r *InventoryRepositoryPGX) GetInventoryByPriceID(
	ctx context.Context,
	priceID string,
) (*domain.Inventory, error) {
	var inv domain.Inventory

	err := r.pool.QueryRow(ctx,
		`SELECT id, product_price_id, name, total_quantity, sold_quantity, status,
		        created_at, updated_at
		 FROM inventories WHERE product_price_id = $1`, priceID,
	).Scan(&inv.ID, &inv.ProductPriceID, &inv.Name, &inv.TotalQuantity,
		&inv.SoldQuantity, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("getting inventory by price ID: %w", err)
	}

	return &inv, nil
}

func (r *InventoryRepositoryPGX) ListInventories(
	ctx context.Context,
) ([]domain.Inventory, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_price_id, name, total_quantity, sold_quantity, status,
		        created_at, updated_at
		 FROM inventories
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing inventories: %w", err)
	}
	defer rows.Close()

	var inventories []domain.Inventory
	for rows.Next() {
		var inv domain.Inventory
		err := rows.Scan(&inv.ID, &inv.ProductPriceID, &inv.Name, &inv.TotalQuantity,
			&inv.SoldQuantity, &inv.Status, &inv.CreatedAt, &inv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning inventory row: %w", err)
		}
		inventories = append(inventories, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating inventory rows: %w", err)
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
	status := inventory.Status
	if status == "" {
		status = "銷售中"
	}

	ct, err := r.pool.Exec(ctx,
		`UPDATE inventories
		 SET name = $1, total_quantity = $2, sold_quantity = $3, status = $4, updated_at = now()
		 WHERE id = $5`,
		inventory.Name, inventory.TotalQuantity, inventory.SoldQuantity,
		status, inventory.ID,
	)
	if err != nil {
		return fmt.Errorf("updating inventory: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrInventoryNotFound
	}

	err = r.pool.QueryRow(ctx,
		`SELECT updated_at FROM inventories WHERE id = $1`, inventory.ID,
	).Scan(&inventory.UpdatedAt)
	if err != nil {
		return fmt.Errorf("reading updated inventory timestamp: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) DeleteInventory(
	ctx context.Context,
	id string,
) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM inventories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting inventory: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrInventoryNotFound
	}

	return nil
}

func (r *InventoryRepositoryPGX) CreateItem(
	ctx context.Context,
	item *domain.InventoryItem,
) error {
	var cost *float64
	if item.Cost != 0 {
		cost = &item.Cost
	}

	dateAdded := item.DateAdded
	if dateAdded == "" {
		dateAdded = time.Now().Format("2006-01-02")
	}

	status := item.Status
	if status == "" {
		status = "可用"
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO inventory_items (inventory_id, item_code, status, cost, date_added)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, status_updated_at, created_at, updated_at`,
		item.InventoryID, item.ItemCode, status, cost, dateAdded,
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
		cost     *float64
		dateAddr time.Time
	)

	err := r.pool.QueryRow(ctx,
		`SELECT id, inventory_id, item_code, status, cost, date_added,
		        status_updated_at, created_at, updated_at
		 FROM inventory_items WHERE id = $1`, id,
	).Scan(&item.ID, &item.InventoryID, &item.ItemCode, &item.Status,
		&cost, &dateAddr, &item.StatusUpdatedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryItemNotFound
		}
		return nil, fmt.Errorf("getting inventory item by ID: %w", err)
	}

	if cost != nil {
		item.Cost = *cost
	}
	item.DateAdded = dateAddr.Format("2006-01-02")

	return &item, nil
}

func (r *InventoryRepositoryPGX) ListItemsByInventoryID(
	ctx context.Context,
	inventoryID string,
) ([]domain.InventoryItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, inventory_id, item_code, status, cost, date_added,
		        status_updated_at, created_at, updated_at
		 FROM inventory_items
		 WHERE inventory_id = $1
		 ORDER BY created_at DESC`, inventoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing inventory items: %w", err)
	}
	defer rows.Close()

	var items []domain.InventoryItem
	for rows.Next() {
		var (
			item     domain.InventoryItem
			cost     *float64
			dateAddr time.Time
		)
		err := rows.Scan(&item.ID, &item.InventoryID, &item.ItemCode, &item.Status,
			&cost, &dateAddr, &item.StatusUpdatedAt,
			&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning inventory item row: %w", err)
		}
		if cost != nil {
			item.Cost = *cost
		}
		item.DateAdded = dateAddr.Format("2006-01-02")
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating inventory item rows: %w", err)
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
	var cost *float64
	if item.Cost != 0 {
		cost = &item.Cost
	}

	dateAdded := item.DateAdded
	if dateAdded == "" {
		dateAdded = time.Now().Format("2006-01-02")
	}

	ct, err := r.pool.Exec(ctx,
		`UPDATE inventory_items
		 SET item_code = $1, status = $2, cost = $3, date_added = $4,
		     status_updated_at = now(), updated_at = now()
		 WHERE id = $5`,
		item.ItemCode, item.Status, cost, dateAdded, item.ID,
	)
	if err != nil {
		return fmt.Errorf("updating inventory item: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrInventoryItemNotFound
	}

	err = r.pool.QueryRow(ctx,
		`SELECT status_updated_at, updated_at FROM inventory_items WHERE id = $1`, item.ID,
	).Scan(&item.StatusUpdatedAt, &item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("reading updated item timestamps: %w", err)
	}

	return nil
}

func (r *InventoryRepositoryPGX) DeleteItem(
	ctx context.Context,
	id string,
) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM inventory_items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting inventory item: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrInventoryItemNotFound
	}

	return nil
}
