# Product Variants Decision

Use a product variant layer for size, color, capacity, and model attributes. Store option name and value in one `product_options` table, associate options through `product_variant_options`, and let each variant reference a shared `product_prices` row. Inventory will belong to variants so equal-priced variants remain independently stockable. SKU is retained as an optional, product-scoped identifier for future ordering and fulfillment features.
