# 2026-08-08 — Category management (categories table + product FK)

## What changed
- New `categories` table (name UNIQUE). Backfilled by existing distinct
  `products.category` values.
- Replaced `products.category` text column with `products.category_id UUID
  REFERENCES categories(id) ON DELETE RESTRICT`; empty `''` values become
  uncategorized.
- Product SQL functions re-created: `create_product` / `update_product` take a
  `category_id`; `list_products` / `get_product_by_id` LEFT JOIN categories and
  return the category name.
- New category CRUD: `create_category`, `list_categories`, `update_category`,
  `delete_category` (raises `R0010` when still referenced).
- APIs `POST /api/categories`, `POST /api/categories/{id}/update`,
  `POST /api/categories/{id}/delete`, `GET /api/categories`. Any authenticated
  member can manage categories for now (per owner; admin gate deferred to a
  future permissions system).

## Why
Admins manage a closed set of categories; product create/edit selects from that
set (JOIN) instead of free-typing, preventing typos and uncontrolled categories.

## Implementation notes (backend complete)
- Migration `016_create_categories.up.sql`: `categories` table, backfill of
  legacy distinct `products.category`, `products.category_id UUID REFERENCES
  categories(id) ON DELETE RESTRICT`, backfill of ids, index, `DROP COLUMN
  products.category`, DROP+CREATE of the four product functions (return shape
  changed), and new category CRUD functions.
- Deviation from the original SQL sketch: `list_products`/`get_product_by_id`
  return `category_id` as a coalesced non-null `VARCHAR`
  (`COALESCE(p.category_id::text, '')`) instead of a nullable `UUID`. Uncategorized
  products would otherwise make pgx fail scanning `NULL` into the `string`
  `Product.CategoryID` (`cannot scan NULL into *string`). This keeps the API
  contract from migration 001 (category was `NOT NULL DEFAULT ''`).
- Deviation: product functions use unqualified column names in `SET` clauses
  (matching migration 006); `SET products.name = ...` is not valid in Postgres
  (`column "products" does not exist`).
- Repos: product repo passes/scans `CategoryID` (empty -> NULL) and the joined
  name; new `src/database/category_repo.go` maps `23505` -> `ErrCategoryNameExists`
  and `R0010` -> `ErrCategoryInUse`.
- Handler `src/api/category_handler.go`; `RegisterCategoryRoutes` uses only
  `AuthMiddleware`; wired in `main.go`.
- Tests: category repo + handler integration tests; product repo/handler/domain
  tests updated to seed and assert `CategoryID`; harness migration lists / drop
  lists / seed truncate list updated.

## Status
Backend done and green (`go build`, `go vet`, integration tests). Frontend + E2E
remain.