CREATE TABLE inventories (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_price_id  UUID        NOT NULL UNIQUE REFERENCES product_prices(id) ON DELETE CASCADE,
    name              VARCHAR(255) NOT NULL,
    total_quantity    INT         NOT NULL DEFAULT 0,
    sold_quantity     INT         NOT NULL DEFAULT 0,
    status            VARCHAR(50) NOT NULL DEFAULT '銷售中',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_id      UUID        NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
    item_code         VARCHAR(255) NOT NULL,
    status            VARCHAR(50) NOT NULL DEFAULT '可用',
    cost              NUMERIC(10,2),
    date_added        DATE        NOT NULL DEFAULT CURRENT_DATE,
    status_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_items_inventory_id ON inventory_items(inventory_id);
