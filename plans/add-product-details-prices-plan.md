# Add Product Details & Prices CRUD Plan

## Goal

Add create endpoints for `product_details` and `product_prices` (nested under products).

## Entrypoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `POST` | `/api/products/{id}/details` | 新增 detail 至指定 product | Required |
| `POST` | `/api/products/{id}/details/{did}/prices` | 新增 price 至指定 detail | Required |

## Request / Response

### CreateDetailRequest
```json
{
  "detail_type": "return_policy",
  "title": "退換貨政策",
  "content": "七天鑑賞期...",
  "sort_order": 1
}
```

### CreatePriceRequest
```json
{
  "label": "成人票",
  "amount": 100.00,
  "currency": "TWD",
  "sort_order": 1
}
```

### Response
```json
{
  "detail": {
    "id": "uuid",
    "product_id": "uuid",
    "detail_type": "return_policy",
    "title": "退換貨政策",
    "content": "七天鑑賞期...",
    "sort_order": 1,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

## Implementation Order

1. Add domain structs (`ProductDetail`, `ProductPrice`) and request/response types
2. Add error sentinels (`ErrDetailNotFound`, `ErrPriceNotFound`, `ErrProductNotFound` already exists)
3. Add `CreateDetail` / `CreatePrice` to `ProductRepository` interface
4. Implement `CreateDetail` / `CreatePrice` in `ProductRepositoryPGX`
5. Add handler methods `CreateDetail` / `CreatePrice`
6. Register routes in router
7. Add database integration tests
8. Add API handler tests
9. Run lint + typecheck + tests
