# Product Manager API

RESTful product and member management API built with Go, gorilla/mux, and PostgreSQL.

## Getting Started

### Prerequisites

- Go 1.25.3+
- Docker (for PostgreSQL)

### Run

```bash
# Start PostgreSQL
docker compose up -d

# Run all migrations
make migrate

# Start server
go run main.go
```

Server starts on `:8080`.

### Test

```bash
# Unit tests
go test ./src/test/domain/...

# Integration tests (requires test database running)
go test -tags=integration ./src/test/...
```

## API

### Products

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/products` | CreateProduct |
| `GET` | `/api/products` | ListProducts |
| `GET` | `/api/products/{id}` | GetProduct |
| `POST` | `/api/products/{id}/update` | UpdateProduct |
| `POST` | `/api/products/{id}/delete` | DeleteProduct |

### Members

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/members/register` | RegisterMember |
| `POST` | `/api/members/login` | LoginMember |
| `POST` | `/api/members/logout` | LogoutMember |
| `GET` | `/api/members/me` | GetCurrentMember |
| `POST` | `/api/members/update` | UpdateMember |

Authentication is handled via HttpOnly `session_key` cookies (valid for 1 day).

## Schema

### Products

Three-tier product model:

```
products 1──N product_details 1──N product_prices
```

- **products** — type, name, status, description, category
- **product_details** — info sections (return_policy, usage_instructions, specifications, shipping_info)
- **product_prices** — price variants with multi-currency support (default TWD)

### Members

```
members 1──N sessions
```

- **members** — id, email (unique), password (bcrypt), name, timestamps
- **sessions** — id, member_id (FK), session_key, expires_at (1 day)

## Project Layout

```
src/
├── api/          # HTTP handlers + router
├── domain/       # Structs + repository interfaces + errors
├── database/     # PostgreSQL repository implementations (pgx)
└── test/         # Unit and integration tests
db/
└── migrations/   # SQL migrations
```
