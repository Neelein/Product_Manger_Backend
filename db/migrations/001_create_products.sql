CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        VARCHAR(50)  NOT NULL,
    name        VARCHAR(255) NOT NULL,
    status      VARCHAR(50)  NOT NULL DEFAULT 'active',
    category    VARCHAR(100),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE product_details (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id         UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    introduction       TEXT,
    usage_instructions TEXT,
    return_policy      TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE product_prices (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_detail_id  UUID        NOT NULL REFERENCES product_details(id) ON DELETE CASCADE,
    label              VARCHAR(255) NOT NULL,
    amount             NUMERIC(10,2) NOT NULL,
    currency           VARCHAR(3)  NOT NULL DEFAULT 'TWD',
    sort_order         INT         NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_details_product_id ON product_details(product_id);
CREATE INDEX idx_product_prices_detail_id   ON product_prices(product_detail_id);
