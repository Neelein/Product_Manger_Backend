# Product Manager API

RESTful product management API built with Go, gorilla/mux, and PostgreSQL.

## Getting Started

### Prerequisites

- Go 1.25.3+
- Docker (for PostgreSQL)

### Run

```bash
# Start PostgreSQL
docker compose up -d

# Run migrations
docker exec -i pm_postgres psql -U root -d productdb < db/migrations/001_create_products.sql

# Start server
go run main.go
```

Server starts on `:8080`.

### Test

```bash
go test ./...
```

## API

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/products` | CreateProduct |
| `GET` | `/api/products` | ListProducts |
| `GET` | `/api/products/{id}` | GetProduct |
| `POST` | `/api/products/{id}/update` | UpdateProduct |
| `POST` | `/api/products/{id}/delete` | DeleteProduct |

## Schema

Three-tier product model:

```
products 1──N product_details 1──N product_prices
```

- **products** — type, name, status, description, category
- **product_details** — info sections (return_policy, usage_instructions, specifications, shipping_info)
- **product_prices** — price variants with multi-currency support (default TWD)

## Project Layout

```
src/
├── api/          # HTTP handlers + router
├── domain/       # Product struct + repository interface
├── database/     # In-memory repo (pgx placeholder)
└── test/         # Tests
db/
└── migrations/   # SQL migrations
```
# Product_Manger_Backend
