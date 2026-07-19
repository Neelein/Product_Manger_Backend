# Product Manager API

RESTful product, inventory, and member management API built with Go, gorilla/mux, and PostgreSQL.

## Tech Stack

- **Language:** Go 1.25.3+
- **Router:** gorilla/mux v1.8.1
- **Database:** PostgreSQL 16 (via pgx/v5)
- **Auth:** Session-based with device fingerprint rotation
- **Infra:** Docker Compose (PostgreSQL)
- **Test:** testify + integration tests with test database

## Getting Started

### Prerequisites

- Go 1.25.3+
- Docker (for PostgreSQL)

### Setup

```bash
# 1. Start PostgreSQL
docker compose up -d

# 2. Apply migrations manually (or via test DB setup)
# Run each migration in db/migrations/ against the database:
psql -h localhost -U root -d productdb -f db/migrations/001_create_products.sql
# ... repeat for 002 through 009

# 3. Start server
DATABASE_URL=postgres://root:root123@localhost:5432/productdb?sslmode=disable go run main.go
```

Server starts on `:8080`. Health check: `GET /api/health`.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `TEST_DATABASE_URL` | — | Test database connection string |

A `dotenv.env` file with defaults is provided.

### Test

```bash
# Unit tests
make test

# Integration tests (requires test database running)
make test-integration
```

## Architecture

### Domain Layers

```
src/
├── api/          # HTTP handlers, middleware, router
├── domain/       # Structs, repository interfaces, sentinel errors
├── database/     # pgx repository implementations
└── test/         # Unit + integration tests
db/
└── migrations/   # SQL migrations (1-9)
```

### Design Patterns

- **Repository pattern** — interfaces in `domain/`, implementations in `database/`
- **Stored functions** — all CRUD via PostgreSQL functions (migration 008), called as `SELECT * FROM function_name(...)`
- **Session auth** — in-memory session cache with TTL; `X-Session-Key` + `X-Device-Fingerprint` headers; rotation on each request
- **Computed fields** — inventory `total_quantity` / `sold_quantity` computed via SQL LEFT JOIN + GROUP BY

## Schema

### Products (3-tier)

```
products 1──1 product_details 1──N product_prices
```

- **products** — name, status, price, category
- **product_details** — introduction, usage_instructions, return_policy
- **product_prices** — price variants with multi-currency, sort_order

### Inventory

```
product_prices 1──1 inventories 1──N inventory_items
```

- **inventories** — linked to product_price_id (UNIQUE), computed name as `{product name}-{price label}`
- **inventory_items** — individual units with item_code, status, cost, date_added

### Members

```
members 1──N sessions
```

- **members** — email (unique), password, name
- **sessions** — session_key, device_fingerprint, expires_at (24h)

## API

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Home |
| GET | `/api/health` | Health check |
| POST | `/api/members/register` | Register member |
| POST | `/api/members/login` | Login (returns session) |
| POST | `/api/members/logout` | Logout |

### Products

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/products` | — | List products |
| POST | `/api/products` | Yes | Create product |
| GET | `/api/products/{productId}` | — | Get product |
| POST | `/api/products/{productId}/update` | — | Update product |
| POST | `/api/products/{productId}/delete` | — | Delete product |
| GET | `/api/products/{productId}/detail` | — | Get product detail |
| POST | `/api/products/{productId}/detail/update` | Yes | Update detail |
| POST | `/api/products/{productId}/details` | Yes | Create detail |
| GET | `/api/products/{productId}/detail/prices` | — | List prices |
| GET | `/api/products/{productId}/detail/prices/{priceId}` | — | Get price |
| POST | `/api/products/{productId}/detail/prices/{priceId}/update` | Yes | Update price |
| POST | `/api/products/{productId}/details/{detailId}/prices` | Yes | Create price |

### Inventory

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/inventories` | — | List inventories |
| POST | `/api/inventories` | Yes | Create inventory |
| GET | `/api/inventories/{inventoryId}` | — | Get inventory (includes product_detail_id) |
| POST | `/api/inventories/{inventoryId}/update` | Yes | Update inventory |
| POST | `/api/inventories/{inventoryId}/delete` | Yes | Delete inventory |
| GET | `/api/inventories/{inventoryId}/items` | — | List items |
| POST | `/api/inventories/{inventoryId}/items` | Yes | Create item |
| GET | `/api/inventories/{inventoryId}/items/{itemId}` | — | Get item |
| POST | `/api/inventories/{inventoryId}/items/{itemId}/update` | Yes | Update item |
| POST | `/api/inventories/{inventoryId}/items/{itemId}/delete` | Yes | Delete item |

### Members

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/members/register` | — | Register |
| POST | `/api/members/login` | — | Login |
| POST | `/api/members/logout` | — | Logout |
| GET | `/api/members/me` | Yes | Get current member |
| POST | `/api/members/update` | Yes | Update profile |

### Auth

Authenticated endpoints require two headers:
- `X-Session-Key` — session token (returned from login)
- `X-Device-Fingerprint` — device identifier

Sessions rotate on each authenticated request (old key becomes invalid, new key returned).

## Migrations

| # | File | Description |
|---|------|-------------|
| 1 | `001_create_products.sql` | products, product_details, product_prices tables |
| 2 | `002_create_members.sql` | members table |
| 3 | `003_add_member_id_to_products.sql` | member_id FK on products |
| 4 | `004_alter_product_details.sql` | Redesign product_details columns |
| 5 | `005_remove_products_description.sql` | Drop description from products |
| 6 | `006_create_inventory.sql` | inventories, inventory_items tables |
| 7 | `007_simplify_inventories.sql` | Drop redundant inventory columns |
| 8 | `008_create_functions.sql` | All CRUD stored functions |
| 9 | `009_add_inventory_id_to_price_functions.sql` | Add inventory_id to price functions |
