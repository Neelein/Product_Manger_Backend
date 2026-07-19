# Products CRUD API Plan

## Entrypoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/products` | 新增產品 |
| `GET` | `/api/products` | 取得產品列表 |
| `GET` | `/api/products/{id}` | 取得單一產品 |
| `POST` | `/api/products/{id}/update` | 修改產品 |
| `POST` | `/api/products/{id}/delete` | 刪除產品 |

## Conventions

- **GET** — read only
- **POST** — all write operations (create, update, delete)
- Route prefix: `/api/products` (follows existing `/api/health` style)
- JSON request/response
- Data store: in-memory first, DB later
