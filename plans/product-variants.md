# Product Variants

## Goal

Support product variants such as clothing sizes, colors, phone capacities, and model numbers while allowing variants with the same price to share one `product_prices` row.

## Data model

- Add `product_options` with `name` and `value` under a product detail.
- Add `product_variants` with optional SKU, status, and a product price reference.
- Add `product_variant_options` to associate a variant with its option values.
- Associate inventory with a product variant rather than only a price.
- Preserve existing price and inventory data by creating a default variant for legacy price rows.

## Backend

- Add migrations, domain types, repository methods, SQL functions, routes, and handlers for options and variants.
- Validate ownership across product, detail, price, option, and variant relationships.
- Enforce unique option values per detail, unique SKU per product when present, and unique option-type usage per variant.
- Keep SKU nullable for now to support future ordering and fulfillment integration.

## Frontend

- Add option and variant API types and client functions.
- Add option and variant management to the product detail page.
- Display each variant's options, SKU, price, and inventory.
- Route inventory creation and management through a variant.

## Verification

- Add backend unit/integration tests for persistence, validation, legacy migration behavior, and variant inventory.
- Add Playwright coverage for creating options, variants sharing a price, and separate variant inventory.
- Review fixes must include API coverage for variant inventory IDs, cross-product mutation rejection, and duplicate combinations on update.
