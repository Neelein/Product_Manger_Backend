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

# 2. Apply migrations with the supported migration runner
DATABASE_URL=postgres://root:root123@localhost:5432/productdb?sslmode=disable go run . migrate

# 3. Start server
DATABASE_URL=postgres://root:root123@localhost:5432/productdb?sslmode=disable go run main.go
```

Server starts on `:8080`. Health check: `GET /api/health`.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | — | PostgreSQL connection string |

A `dotenv.env` file with defaults is provided.

### Test

```bash
# Unit tests
make test

# Safe integration tests (uses a random schema inside productdb; public data is never reset)
DATABASE_URL=postgres://root:root123@localhost:5432/productdb?sslmode=disable make test-integration-productdb
```

CI provisions PostgreSQL with the `productdb` database, runs `go run . migrate`
before integration tests, and passes the same `DATABASE_URL` explicitly. The CI
credentials are disposable service-container credentials, not application or
repository secrets.

Integration tests do not use `TEST_DATABASE_URL`. Each run creates a random schema,
configures every test pool connection with that schema first in `search_path`, and
drops only that schema during cleanup. The public `root@gmail.com` member is checked
as a sentinel before and after the run. Do not interrupt the process while it is
creating the schema; an orphaned `integration_*` schema is harmless and can be
reported for manual review rather than resetting productdb.

### Storefront browser E2E fixture

The repeatable fixture command reads `DATABASE_URL` and defaults to the dedicated
database `productdb_storefront_e2e`. Apply migrations 001 through 022 first, and
create that database if it does not exist. The command only upserts the fixed
category, product, detail, and two prices; it does not truncate other data. It
also verifies the resulting rows before exiting successfully.

```bash
DATABASE_URL=postgres://root:root123@localhost:5432/productdb_storefront_e2e?sslmode=disable \
  go run ./src/test/seed_storefront
```

The storefront E2E harness should invoke the command from `backend/` exactly as
above, then start the API against the same `DATABASE_URL`.

### Checkout order contract

`POST /api/orders` accepts `items`, `customer` (`name`, `phone`, `email`), and
`delivery_method` (`email` or `home_address`). `shipping_address.address` is
required only for `home_address`. The order response retains the existing
`customer_snapshot` and `shipping_address_snapshot` fields, now containing JSON
objects shaped as `{name, phone, email}` and `{delivery_method, address?}`.
Payment remains a separate optional API flow; order creation does not require it.

## Architecture

### Domain Layers

```
src/
├── api/          # HTTP handlers, middleware, router
├── domain/       # Structs, repository interfaces, sentinel errors
├── database/     # pgx repository implementations
└── test/         # Unit + integration tests
db/
└── migrations/   # SQL migrations (001-022, .up.sql)
```

### Design Patterns

- **Repository pattern** — interfaces in `domain/`, implementations in `database/`
- **Stored functions** — all CRUD via PostgreSQL functions (migration 006), called as `SELECT * FROM function_name(...)`
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
product_prices 1──N product_variants 1──N inventories 1──N inventory_items
```

Price responses do not include `inventory_id`. Inventory is related to a price
through `product_variant_id` and must be queried from the inventory endpoints;
`InventoryID` remains part of inventory and inventory-item responses.

- **inventories** — linked to product variants through `product_variant_id`, computed name as `{product name}-{price label}`
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
| GET | `/api/products` | — | List products; optional `keyword` (case-insensitive name match) and `category_id` (exact UUID match) query parameters |
| POST | `/api/products` | Yes | Create product |
| GET | `/api/products/{productId}` | — | Get product |
| POST | `/api/products/{productId}/update` | — | Update product |
| POST | `/api/products/{productId}/delete` | — | Delete product |
| GET | `/api/products/{productId}/images` | — | List product images as `{ "images": [...] }` |
| POST | `/api/products/{productId}/images` | Employee | Multipart upload with repeated `images` fields; JPEG/PNG/WebP, 10 MB per file, maximum three persisted images total |
| GET | `/media/images/products/{productId}/{filename}` | — | Public product image retrieval |
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

Inventory creation requires `product_variant_id` (and optionally `status`). `CreateInventory`
does not use the legacy `product_price_id` request field; unknown JSON fields, including
`product_price_id`, are currently ignored. `product_price_id` may still appear in inventory
responses as a value derived through the variant relationship.

## Migrations

| # | File | Description |
|---|------|-------------|
| 001 | `001_create_products.up.sql` | products, product_details, product_prices tables |
| 002 | `002_create_members.up.sql` | members table |
| 003 | `003_add_member_id_to_products.up.sql` | member_id FK on products |
| 004 | `004_create_inventory.up.sql` | inventories, inventory_items tables |
| 005 | `005_simplify_inventories.up.sql` | Simplify inventory columns |
| 006 | `006_create_functions.up.sql` | CRUD stored functions |
| 007 | `007_add_inventory_id_to_price_functions.up.sql` | Define price functions without a direct inventory column |
| 008 | `008_create_announcements.up.sql` | announcements table and functions |
| 009 | `009_set_not_null.up.sql` | Set required columns to NOT NULL |
| 010 | `010_create_chat.up.sql` | chat rooms, messages, memberships, and receipts |
| 011 | `011_list_members.up.sql` | Member listing function |
| 012 | `012_create_events.up.sql` | events and event viewers |
| 013 | `013_add_monthly_filters.up.sql` | Monthly event filters |
| 014 | `014_insert_admin_member.up.sql` | Seed administrator member |
| 015 | `015_create_registration_codes.up.sql` | registration_codes table and functions |
| 016 | `016_create_categories.up.sql` | categories table and functions |
| 017 | `017_create_product_variants.up.sql` | product variants and variant options; link inventories to variants |
| 018 | `018_members_orders.up.sql` | orders, order items, and status history |
| 019 | `019_create_payments.up.sql` | payments table and functions |
| 020 | `020_create_product_images.up.sql` | product image metadata and functions |
| 021 | `021_link_price_inventory_by_variant.up.sql` | Link inventory and price lookups through product variants |
| 022 | `022_repair_inventory_function_contract.up.sql` | Repair inventory function return contracts |
