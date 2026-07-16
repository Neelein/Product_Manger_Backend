# Architecture Plan

## Directory structure

```
src/
  ├── api/
  │   ├── handler.go          # HTTP handlers — Parse request, call domain, write response
  │   └── router.go           # Register product routes on mux.Router
  ├── domain/
  │   ├── product.go          # Product struct
  │   └── repository.go       # ProductRepository interface (port)
  ├── database/
  │   ├── db.go               # (placeholder for pgx connection later)
  │   ├── inmemory_repo.go    # In-memory ProductRepository implementation
  │   └── product_repo.go     # (placeholder for pgx implementation later)
  └── test/
      ├── domain/
      │   └── product_test.go        # Test Product struct
      ├── database/
      │   └── inmemory_repo_test.go  # Test in-memory CRUD
      └── api/
          └── handler_test.go        # Test HTTP endpoints
```

## Data flow

```
HTTP Request → api/handler → domain/repository interface → database/repo → in-memory map
```

## Component details

### domain/product.go

```go
type Product struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Category    string    `json:"category"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### domain/repository.go

```go
type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    List(ctx context.Context) ([]Product, error)
    GetByID(ctx context.Context, id string) (*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id string) error
}
```

### database/inmemory_repo.go

- Use `sync.RWMutex` + `map[string]Product`
- Generate UUID for product ID
- Implement all 5 CRUD methods from ProductRepository interface

### api/handler.go

- Receive `ProductRepository` via dependency injection
- Parse JSON request body
- Call domain repository methods
- Encode JSON response (standard response wrapper)

### api/router.go

```go
func RegisterProductRoutes(r *mux.Router, repo domain.ProductRepository)
```

### main.go

- Create in-memory repo
- Call RegisterProductRoutes
- Existing home + health handlers remain

## Entrypoints

| Method | Path | Handler |
|--------|------|---------|
| `POST` | `/api/products` | CreateProduct |
| `GET` | `/api/products` | ListProducts |
| `GET` | `/api/products/{id}` | GetProduct |
| `POST` | `/api/products/{id}/update` | UpdateProduct |
| `POST` | `/api/products/{id}/delete` | DeleteProduct |
