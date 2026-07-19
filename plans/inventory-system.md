# Plan — Inventory System (2026-07-18)

## Overview

Add inventory tracking system with two tables: `inventories` (summary) and `inventory_items` (individual items). Each `inventory` is linked one-to-one with `product_prices`, and each `inventory` has many `inventory_items`.

## Schema

- **inventories**: id, product_price_id (FK UNIQUE), name, total_quantity, sold_quantity, status (完售/註銷/銷售中), created_at, updated_at
- **inventory_items**: id, inventory_id (FK), item_code, status (出售/註銷/可用), cost, date_added, status_updated_at, created_at, updated_at

## Implementation Order

1. Migration: `db/migrations/006_create_inventory.sql`
2. Domain structs: `src/domain/inventory.go`
3. Sentinel errors: `src/domain/errors.go`
4. Repository interface: `src/domain/repository.go`
5. Repository impl: `src/database/inventory_repo.go`
6. Handler: `src/api/inventory_handler.go`
7. Routes: `src/api/router.go`
8. Wiring: `main.go`
9. Database integration tests: `src/test/database/inventory_repo_test.go`
10. Handler integration tests: `src/test/api/inventory_handler_test.go`

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/inventories | Yes | Create inventory |
| GET | /api/inventories | No | List inventories |
| GET | /api/inventories/{id} | No | Get inventory |
| POST | /api/inventories/{id}/update | Yes | Update inventory |
| POST | /api/inventories/{id}/delete | Yes | Delete inventory |
| POST | /api/inventories/{id}/items | Yes | Create inventory item |
| GET | /api/inventories/{id}/items | No | List items |
| GET | /api/inventories/{id}/items/{iid} | No | Get item |
| POST | /api/inventories/{id}/items/{iid}/update | Yes | Update item |
| POST | /api/inventories/{id}/items/{iid}/delete | Yes | Delete item |
