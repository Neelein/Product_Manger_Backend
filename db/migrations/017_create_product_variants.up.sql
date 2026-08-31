CREATE TABLE product_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_detail_id UUID NOT NULL REFERENCES product_details(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_detail_id, name, value)
);

CREATE INDEX idx_product_options_detail_id ON product_options(product_detail_id);

CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_detail_id UUID NOT NULL REFERENCES product_details(id) ON DELETE CASCADE,
    product_price_id UUID NOT NULL REFERENCES product_prices(id) ON DELETE RESTRICT,
    sku VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_detail_id, sku)
);

CREATE UNIQUE INDEX idx_product_variants_sku
    ON product_variants (product_detail_id, sku) WHERE sku IS NOT NULL;
CREATE INDEX idx_product_variants_price_id ON product_variants(product_price_id);

CREATE TABLE product_variant_options (
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES product_options(id) ON DELETE CASCADE,
    PRIMARY KEY (variant_id, option_id)
);

CREATE INDEX idx_variant_options_option_id ON product_variant_options(option_id);

ALTER TABLE inventories ADD COLUMN product_variant_id UUID;
ALTER TABLE inventories DROP CONSTRAINT inventories_product_price_id_key;

INSERT INTO product_variants (product_detail_id, product_price_id, status)
SELECT product_detail_id, id, 'active' FROM product_prices;

UPDATE inventories i
SET product_variant_id = v.id
FROM product_variants v
WHERE v.product_price_id = i.product_price_id;

ALTER TABLE inventories
    ALTER COLUMN product_variant_id SET NOT NULL,
    ADD CONSTRAINT inventories_product_variant_id_fkey
        FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX idx_inventories_variant_id ON inventories(product_variant_id);

CREATE OR REPLACE FUNCTION create_product_price(
    p_product_detail_id UUID, p_label VARCHAR, p_amount NUMERIC, p_currency VARCHAR, p_sort_order INT
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
DECLARE v_price_id UUID;
BEGIN
    INSERT INTO product_prices(product_detail_id, label, amount, currency, sort_order)
    VALUES (p_product_detail_id, p_label, p_amount, COALESCE(p_currency, 'TWD'), p_sort_order)
    RETURNING product_prices.id INTO v_price_id;
    INSERT INTO product_variants(product_detail_id, product_price_id, status)
    VALUES (p_product_detail_id, v_price_id, 'active');
    RETURN QUERY SELECT pp.id, pp.created_at, pp.updated_at FROM product_prices pp WHERE pp.id = v_price_id;
END; $$;

DROP FUNCTION IF EXISTS get_product_price_by_id(UUID);
DROP FUNCTION IF EXISTS list_product_prices_by_detail(UUID);
DROP FUNCTION IF EXISTS create_inventory(UUID, VARCHAR);
DROP FUNCTION IF EXISTS get_inventory_by_id(UUID);
DROP FUNCTION IF EXISTS get_inventory_by_price_id(UUID);
DROP FUNCTION IF EXISTS list_inventories();

CREATE OR REPLACE FUNCTION create_product_option(p_detail_id UUID, p_name VARCHAR, p_value VARCHAR)
RETURNS TABLE(id UUID, product_detail_id UUID, name VARCHAR, value VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO product_options(product_detail_id, name, value)
    VALUES (p_detail_id, p_name, p_value)
    RETURNING id, product_detail_id, name, value, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_product_option_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, name VARCHAR, value VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$ SELECT id, product_detail_id, name, value, created_at, updated_at FROM product_options WHERE id = p_id; $$;

CREATE OR REPLACE FUNCTION list_product_options(p_detail_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, name VARCHAR, value VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$ SELECT id, product_detail_id, name, value, created_at, updated_at FROM product_options WHERE product_detail_id = p_detail_id ORDER BY name, value; $$;

CREATE OR REPLACE FUNCTION update_product_option(p_id UUID, p_name VARCHAR, p_value VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$ UPDATE product_options SET name = p_name, value = p_value, updated_at = now() WHERE id = p_id RETURNING updated_at; $$;

CREATE OR REPLACE FUNCTION delete_product_option(p_id UUID) RETURNS BOOLEAN LANGUAGE plpgsql AS $$
BEGIN DELETE FROM product_options WHERE id = p_id; RETURN FOUND; END; $$;

CREATE OR REPLACE FUNCTION create_product_variant(p_detail_id UUID, p_price_id UUID, p_sku VARCHAR, p_status VARCHAR, p_option_ids UUID[])
RETURNS TABLE(id UUID, product_detail_id UUID, product_price_id UUID, sku VARCHAR, status VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
DECLARE v_id UUID;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM product_prices pp WHERE pp.id = p_price_id AND pp.product_detail_id = p_detail_id) THEN RAISE EXCEPTION 'price does not belong to detail' USING ERRCODE = 'R0020'; END IF;
    IF EXISTS (SELECT 1 FROM product_options o WHERE o.id = ANY(COALESCE(p_option_ids, ARRAY[]::UUID[])) AND o.product_detail_id <> p_detail_id) THEN RAISE EXCEPTION 'option does not belong to detail' USING ERRCODE = 'R0021'; END IF;
    IF (SELECT COUNT(*) FROM unnest(COALESCE(p_option_ids, ARRAY[]::UUID[]))) <> (SELECT COUNT(DISTINCT o.name) FROM product_options o WHERE o.id = ANY(COALESCE(p_option_ids, ARRAY[]::UUID[]))) THEN RAISE EXCEPTION 'duplicate option name' USING ERRCODE = 'R0022'; END IF;
    IF NULLIF(p_sku, '') IS NOT NULL AND EXISTS (SELECT 1 FROM product_variants v WHERE v.product_detail_id = p_detail_id AND v.sku = NULLIF(p_sku, '')) THEN RAISE EXCEPTION 'duplicate sku' USING ERRCODE = 'R0023'; END IF;
    IF EXISTS (SELECT 1 FROM product_variants v WHERE v.product_detail_id = p_detail_id AND NOT EXISTS (SELECT 1 FROM product_variant_options x WHERE x.variant_id = v.id AND x.option_id <> ALL(COALESCE(p_option_ids, ARRAY[]::UUID[]))) AND (SELECT COUNT(*) FROM product_variant_options x WHERE x.variant_id = v.id) = COALESCE(array_length(p_option_ids, 1), 0)) THEN RAISE EXCEPTION 'duplicate option combination' USING ERRCODE = 'R0024'; END IF;
    INSERT INTO product_variants AS pv(product_detail_id, product_price_id, sku, status) VALUES (p_detail_id, p_price_id, NULLIF(p_sku, ''), COALESCE(NULLIF(p_status, ''), 'active')) RETURNING pv.id INTO v_id;
    INSERT INTO product_variant_options(variant_id, option_id) SELECT v_id, unnest(COALESCE(p_option_ids, ARRAY[]::UUID[]));
    RETURN QUERY SELECT v.id, v.product_detail_id, v.product_price_id, v.sku, v.status, v.created_at, v.updated_at FROM product_variants v WHERE v.id = v_id;
END; $$;

CREATE OR REPLACE FUNCTION get_product_variant_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, product_price_id UUID, sku VARCHAR, status VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, option_ids UUID[])
LANGUAGE sql AS $$ SELECT v.id, v.product_detail_id, v.product_price_id, v.sku, v.status, v.created_at, v.updated_at, COALESCE(array_agg(x.option_id) FILTER (WHERE x.option_id IS NOT NULL), ARRAY[]::UUID[]) FROM product_variants v LEFT JOIN product_variant_options x ON x.variant_id = v.id WHERE v.id = p_id GROUP BY v.id; $$;

CREATE OR REPLACE FUNCTION list_product_variants(p_detail_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, product_price_id UUID, sku VARCHAR, status VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ, option_ids UUID[])
LANGUAGE sql AS $$ SELECT v.id, v.product_detail_id, v.product_price_id, v.sku, v.status, v.created_at, v.updated_at, COALESCE(array_agg(x.option_id) FILTER (WHERE x.option_id IS NOT NULL), ARRAY[]::UUID[]) FROM product_variants v LEFT JOIN product_variant_options x ON x.variant_id = v.id WHERE v.product_detail_id = p_detail_id GROUP BY v.id ORDER BY v.created_at; $$;

CREATE OR REPLACE FUNCTION update_product_variant(p_id UUID, p_price_id UUID, p_sku VARCHAR, p_status VARCHAR, p_option_ids UUID[])
RETURNS TABLE(updated_at TIMESTAMPTZ) LANGUAGE plpgsql AS $$
DECLARE v_updated_at TIMESTAMPTZ;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM product_variants v JOIN product_prices p ON p.product_detail_id = v.product_detail_id WHERE v.id = p_id AND p.id = p_price_id) THEN RAISE EXCEPTION 'price does not belong to variant' USING ERRCODE = 'R0020'; END IF;
    IF EXISTS (SELECT 1 FROM product_options o WHERE o.id = ANY(COALESCE(p_option_ids, ARRAY[]::UUID[])) AND o.product_detail_id <> (SELECT product_detail_id FROM product_variants WHERE id = p_id)) THEN RAISE EXCEPTION 'option does not belong to detail' USING ERRCODE = 'R0021'; END IF;
    IF (SELECT COUNT(*) FROM unnest(COALESCE(p_option_ids, ARRAY[]::UUID[]))) <> (SELECT COUNT(DISTINCT name) FROM product_options WHERE id = ANY(COALESCE(p_option_ids, ARRAY[]::UUID[]))) THEN RAISE EXCEPTION 'duplicate option name' USING ERRCODE = 'R0022'; END IF;
    IF NULLIF(p_sku, '') IS NOT NULL AND EXISTS (SELECT 1 FROM product_variants v WHERE v.product_detail_id = (SELECT product_detail_id FROM product_variants WHERE id = p_id) AND v.sku = NULLIF(p_sku, '') AND v.id <> p_id) THEN RAISE EXCEPTION 'duplicate sku' USING ERRCODE = 'R0023'; END IF;
    IF EXISTS (
        SELECT 1
        FROM product_variants v
        WHERE v.product_detail_id = (SELECT product_detail_id FROM product_variants WHERE id = p_id)
          AND v.id <> p_id
          AND NOT EXISTS (
              SELECT 1 FROM product_variant_options x
              WHERE x.variant_id = v.id
                AND x.option_id <> ALL(COALESCE(p_option_ids, ARRAY[]::UUID[]))
          )
          AND (SELECT COUNT(*) FROM product_variant_options x WHERE x.variant_id = v.id) = COALESCE(array_length(p_option_ids, 1), 0)
    ) THEN RAISE EXCEPTION 'duplicate option combination' USING ERRCODE = 'R0024'; END IF;
    DELETE FROM product_variant_options WHERE variant_id = p_id;
    INSERT INTO product_variant_options SELECT p_id, unnest(COALESCE(p_option_ids, ARRAY[]::UUID[]));
    UPDATE product_variants AS pv SET product_price_id = p_price_id, sku = NULLIF(p_sku, ''), status = COALESCE(NULLIF(p_status, ''), 'active'), updated_at = now() WHERE pv.id = p_id RETURNING pv.updated_at INTO v_updated_at;
    RETURN QUERY SELECT v_updated_at;
END; $$;

CREATE OR REPLACE FUNCTION delete_product_variant(p_id UUID) RETURNS BOOLEAN LANGUAGE plpgsql AS $$
BEGIN DELETE FROM product_variants WHERE id = p_id; RETURN FOUND; END; $$;

CREATE OR REPLACE FUNCTION create_inventory(p_product_variant_id UUID, p_status VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) LANGUAGE sql AS $$
    INSERT INTO inventories(product_variant_id, product_price_id, status)
    SELECT v.id, v.product_price_id, COALESCE(p_status, '銷售中') FROM product_variants v
    WHERE v.id = p_product_variant_id
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_inventory_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID, product_id UUID, name VARCHAR, status VARCHAR, total_quantity BIGINT, sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) LANGUAGE sql AS $$
 SELECT i.id, i.product_variant_id, i.product_price_id, pr.product_detail_id, pd.product_id, CONCAT(p.name, '-', pr.label)::VARCHAR, i.status, COUNT(it.id), COUNT(it.id) FILTER (WHERE it.status = '出售'), i.created_at, i.updated_at FROM inventories i JOIN product_variants v ON v.id=i.product_variant_id JOIN product_prices pr ON pr.id=v.product_price_id JOIN product_details pd ON pd.id=pr.product_detail_id JOIN products p ON p.id=pd.product_id LEFT JOIN inventory_items it ON it.inventory_id=i.id WHERE i.id=p_id GROUP BY i.id, i.product_variant_id, i.product_price_id, pr.product_detail_id, pd.product_id, p.name, pr.label, i.status, i.created_at, i.updated_at;
$$;
CREATE OR REPLACE FUNCTION get_inventory_by_price_id(p_price_id UUID)
RETURNS TABLE(id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID, product_id UUID, name VARCHAR, status VARCHAR, total_quantity BIGINT, sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$ SELECT * FROM get_inventory_by_id((SELECT i.id FROM inventories i WHERE i.product_price_id=p_price_id LIMIT 1)); $$;
CREATE OR REPLACE FUNCTION list_inventories()
RETURNS TABLE(id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID, product_id UUID, name VARCHAR, status VARCHAR, total_quantity BIGINT, sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$ SELECT x.* FROM inventories i CROSS JOIN LATERAL get_inventory_by_id(i.id) x ORDER BY x.created_at DESC; $$;

CREATE OR REPLACE FUNCTION get_product_price_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) LANGUAGE sql AS $$ SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency, pp.sort_order, pp.created_at, pp.updated_at FROM product_prices pp WHERE pp.id=p_id; $$;
CREATE OR REPLACE FUNCTION list_product_prices_by_detail(p_detail_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) LANGUAGE sql AS $$ SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency, pp.sort_order, pp.created_at, pp.updated_at FROM product_prices pp WHERE pp.product_detail_id=p_detail_id ORDER BY pp.sort_order; $$;
