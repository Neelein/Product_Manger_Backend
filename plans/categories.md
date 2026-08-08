# Category management

Boundary: categories become a first-class table; products reference a category via FK instead of free-text; a category management card lives on the dashboard.

> Status (2026-08-08): Backend implemented and green (`go build`, `go vet`,
> integration tests). Frontend + E2E pending.

## Decisions (confirmed 2026-08-08)
- Entry card on `/home` (DashboardPage grid) + navbar link 「類別」.
- Any authenticated member may create/rename/delete categories for now (admin gate deferred to future permissions system).
- Create + rename + restricted delete (block when still referenced by products).
- Replace `products.category` text with `products.category_id UUID REFERENCES categories(id) ON DELETE RESTRICT`; backfill legacy distinct values; `''` becomes NULL. Product read APIs still return the category name via LEFT JOIN.
- API: `GET /api/categories` (public); `POST /api/categories`, `POST /api/categories/{id}/update`, `POST /api/categories/{id}/delete` (any authenticated member).

## Backend
1. Migration `016_create_categories.up.sql`:
   - `categories(id UUID PK gen_random_uuid, name VARCHAR(100) UNIQUE NOT NULL, created_at, updated_at)`.
   - Backfill distinct non-empty `products.category` → categories.
   - `ADD products.category_id UUID NULL REFERENCES categories(id) ON DELETE RESTRICT`; backfill by name; index; `DROP COLUMN products.category`.
   - DROP+CREATE product functions: `create_product(p_name, p_status, p_category_id UUID, p_member_id)`, `update_product(p_id, p_name, p_status, p_category_id UUID)`, `list_products()`/`get_product_by_id(p_id)` LEFT JOIN categories → include `category_id` + joined `category` name.
   - New functions: `create_category(p_name)` → returns (id,name,created_at,updated_at), `list_categories()`, `update_category(p_id,p_name)` → (updated_at), `delete_category(p_id)` → bool, raises `R0010` when in use.
2. Domain: `Product.CategoryID` added; request types `Category` → `CategoryID`; new `Category` type + `CategoryRepository` interface; sentinel `ErrCategoryInUse`, `ErrCategoryNotFound`, `ErrCategoryNameExists` (with 409/404 mapping).
3. Repo: product repo passes/scans `CategoryID` + joined name; new `src/database/category_repo.go`.
4. API: new `src/api/category_handler.go` (Create 409 dup, Update 409 dup/404, Delete 404 not found / 409 in-use, List); new `RegisterCategoryRoutes` wrapped in `AuthMiddleware` only; wire in `main.go`.
5. Tests: update product repo/handler tests to seed+use a category id; new category repo/handler tests; append `016_create_categories.up.sql` to `runMigration` lists and add `categories` / category functions to `dropTables` in `src/test/api/handler_test.go` and `src/test/database/product_repo_test.go`; update `src/test/seed/seed_test_db.go` truncate list.
6. Verify: `go build ./...`, `go vet ./...`, `go test -tags=integration -count=1 -p=1 ./src/test/...`.

## Frontend
- Types: `Category {id,name,created_at,updated_at}`, `CreateCategoryRequest`, `CategoryResponse`, `CategoryListResponse`; `Product.category_id`; `CreateProductRequest.category_id`.
- `src/api/categories.ts` (list/create/update/delete via POST `.update`/`.delete`), `src/hooks/useCategories.ts`.
- `src/pages/CategoryListPage.tsx`: `.categories-page`-scoped table (RegistrationCodesPage pattern) — create row, inline rename, delete blocked message.
- `DashboardPage.tsx`: add 類別管理 card → `/categories`.
- `App.tsx`: `/categories` route under `ProtectedRoute` (NOT admin).
- `Layout.tsx`: add 「類」 nav link (visible when logged in).
- `ProductForm.tsx`: replace text input with `<select>` fed by `GET /api/categories`, submit `category_id`, initialize from `product.category_id`.
- `ProductCard`/`ProductDetail` unchanged (backend returns name).
- Verify: `npx tsc -b`, `npm run lint`.

## E2E
- `helpers.ts`: `freshCategory()` pg INSERT returns {id,name}.
- `product.spec.ts`/`inventory.spec.ts`: create category, `selectOption` in 分類 select.
- New `e2e/categories.spec.ts`: create + rename category; create category assigned in product form; delete-blocked error when in use; `global-setup` add `categories` table.
- Run: full `npx playwright test`.

## Docs
- `plans/categories.md`, `decisions/2026-08-08-categories.md` (both repos as appropriate).