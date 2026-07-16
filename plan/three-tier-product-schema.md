# Three-Tier Product Schema Plan

## Overview

Split product data into three related tables:
- `products` — main product info (type, name, status)
- `product_details` — info sections (return policy, usage, specs, shipping)
- `product_prices` — price variants (adult/child ticket, size variants)

Hierarchy: `products 1→N product_details 1→N product_prices`

---

## Table Definitions

### `products`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() |
| product_type | VARCHAR(50) | NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| description | TEXT | |
| status | VARCHAR(20) | NOT NULL DEFAULT 'draft' |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

### `product_details`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() |
| product_id | UUID | FK → products(id) ON DELETE CASCADE, NOT NULL |
| detail_type | VARCHAR(50) | NOT NULL |
| title | VARCHAR(255) | NOT NULL |
| content | TEXT | |
| sort_order | INT | NOT NULL DEFAULT 0 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(product_id)`, `(product_id, sort_order)`

`detail_type` values: `return_policy`, `usage_instructions`, `specifications`, `shipping_info`

### `product_prices`

| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() |
| product_detail_id | UUID | FK → product_details(id) ON DELETE CASCADE, NOT NULL |
| name | VARCHAR(255) | NOT NULL |
| price | DECIMAL(12,2) | NOT NULL |
| currency | VARCHAR(3) | NOT NULL DEFAULT 'TWD' |
| sort_order | INT | NOT NULL DEFAULT 0 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT now() |

Indexes: `(product_detail_id)`, `(product_detail_id, sort_order)`

---

## Domain Model

```go
type Product struct {
    ID          string           `json:"id"`
    ProductType string           `json:"product_type"`
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Status      string           `json:"status"`
    Details     []ProductDetail  `json:"details,omitempty"`
    CreatedAt   time.Time        `json:"created_at"`
    UpdatedAt   time.Time        `json:"updated_at"`
}

type ProductDetail struct {
    ID         string          `json:"id"`
    ProductID  string          `json:"product_id"`
    DetailType string          `json:"detail_type"`
    Title      string          `json:"title"`
    Content    string          `json:"content"`
    SortOrder  int             `json:"sort_order"`
    Prices     []ProductPrice  `json:"prices,omitempty"`
    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}

type ProductPrice struct {
    ID               string    `json:"id"`
    ProductDetailID  string    `json:"product_detail_id"`
    Name             string    `json:"name"`
    Price            float64   `json:"price"`
    Currency         string    `json:"currency"`
    SortOrder        int       `json:"sort_order"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

---

## Repository Interface

```go
type ProductRepository interface {
    // Product CRUD
    Create(ctx context.Context, product *Product) error
    List(ctx context.Context) ([]Product, error)
    GetByID(ctx context.Context, id string) (*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id string) error

    // Detail CRUD
    CreateDetail(ctx context.Context, detail *ProductDetail) error
    UpdateDetail(ctx context.Context, detail *ProductDetail) error
    DeleteDetail(ctx context.Context, id string) error

    // Price CRUD
    CreatePrice(ctx context.Context, price *ProductPrice) error
    UpdatePrice(ctx context.Context, price *ProductPrice) error
    DeletePrice(ctx context.Context, id string) error
}
```

---

## API Routes

```
POST   /api/products                           → CreateProduct
GET    /api/products                           → ListProducts
GET    /api/products/{id}                      → GetProduct (nested details + prices)
POST   /api/products/{id}/update               → UpdateProduct
POST   /api/products/{id}/delete               → DeleteProduct

POST   /api/products/{id}/details              → CreateDetail
POST   /api/products/{id}/details/{did}/update → UpdateDetail
POST   /api/products/{id}/details/{did}/delete → DeleteDetail

POST   /api/products/{id}/details/{did}/prices           → CreatePrice
POST   /api/products/{id}/details/{did}/prices/{pid}/update → UpdatePrice
POST   /api/products/{id}/details/{did}/prices/{pid}/delete → DeletePrice
```

---

## Implementation Order

1. SQL migration file (`migrations/001_create_products.sql`)
2. Update domain structs in `src/domain/product.go`
3. Update repository interface in `src/domain/repository.go`
4. Update in-memory repo in `src/database/inmemory_repo.go`
5. Update API handlers in `src/api/handler.go`
6. Update routes in `src/api/router.go`
7. Update request/response types
8. Add tests
