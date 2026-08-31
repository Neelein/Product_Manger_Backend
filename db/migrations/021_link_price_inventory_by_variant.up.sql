-- Migrate legacy rows before removing the denormalized price relationship.
UPDATE inventories i
SET product_variant_id = v.id
FROM product_variants v
WHERE i.product_variant_id IS NULL
  AND v.product_price_id = i.product_price_id;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM inventories i
        LEFT JOIN product_variants v ON v.id = i.product_variant_id
        WHERE v.id IS NULL
           OR i.product_price_id IS DISTINCT FROM v.product_price_id
    ) THEN
        RAISE EXCEPTION 'inventories contain rows whose product_variant_id does not match product_price_id';
    END IF;
END $$;

-- These functions still refer to the legacy column in migrations 006/017.
-- Drop them before dropping the column, then recreate their final contracts.
DROP FUNCTION IF EXISTS create_inventory(UUID, VARCHAR);
DROP FUNCTION IF EXISTS get_inventory_by_id(UUID);
DROP FUNCTION IF EXISTS get_inventory_by_price_id(UUID);
DROP FUNCTION IF EXISTS list_inventories();
DROP FUNCTION IF EXISTS get_product_price_by_id(UUID);
DROP FUNCTION IF EXISTS list_product_prices_by_detail(UUID);

ALTER TABLE inventories
    DROP CONSTRAINT IF EXISTS inventories_product_price_id_fkey,
    DROP CONSTRAINT IF EXISTS inventories_product_price_id_key,
    DROP COLUMN IF EXISTS product_price_id;

CREATE OR REPLACE FUNCTION create_inventory(p_product_variant_id UUID, p_status VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO inventories(product_variant_id, status)
    SELECT v.id, COALESCE(p_status, '銷售中')
    FROM product_variants v
    WHERE v.id = p_product_variant_id
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_inventory_by_id(p_id UUID)
RETURNS TABLE(
    id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID,
    product_id UUID, name VARCHAR, variant_name VARCHAR, status VARCHAR, total_quantity BIGINT,
    sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    WITH variant_options AS (
        SELECT vo.variant_id,
               string_agg(CONCAT(o.name, ': ', o.value), ' / ' ORDER BY o.name, o.value)::VARCHAR AS variant_name
        FROM product_variant_options vo
        JOIN product_options o ON o.id = vo.option_id
        WHERE vo.variant_id = (SELECT product_variant_id FROM inventories WHERE id = p_id)
        GROUP BY vo.variant_id
    ), item_quantities AS (
        SELECT it.inventory_id,
               COUNT(*) AS total_quantity,
               COUNT(*) FILTER (WHERE it.status = '出售') AS sold_quantity
        FROM inventory_items it
        WHERE it.inventory_id = p_id
        GROUP BY it.inventory_id
    )
    SELECT i.id, i.product_variant_id, v.product_price_id, pr.product_detail_id,
           pd.product_id,
           CONCAT_WS('-', p.name, pr.label, NULLIF(vo.variant_name, ''))::VARCHAR,
           COALESCE(vo.variant_name, '')::VARCHAR,
           i.status,
           COALESCE(iq.total_quantity, 0), COALESCE(iq.sold_quantity, 0),
           i.created_at, i.updated_at
    FROM inventories i
    JOIN product_variants v ON v.id = i.product_variant_id
    JOIN product_prices pr ON pr.id = v.product_price_id
    JOIN product_details pd ON pd.id = pr.product_detail_id
    JOIN products p ON p.id = pd.product_id
    LEFT JOIN variant_options vo ON vo.variant_id = v.id
    LEFT JOIN item_quantities iq ON iq.inventory_id = i.id
    WHERE i.id = p_id;
$$;

CREATE OR REPLACE FUNCTION get_inventory_by_price_id(p_product_price_id UUID)
RETURNS TABLE(
    id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID,
    product_id UUID, name VARCHAR, variant_name VARCHAR, status VARCHAR, total_quantity BIGINT,
    sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT x.*
    FROM get_inventory_by_id((
        SELECT i.id
        FROM inventories i
        JOIN product_variants v ON v.id = i.product_variant_id
        WHERE v.product_price_id = p_product_price_id
        ORDER BY v.created_at, v.id, i.created_at, i.id
        LIMIT 1
    )) x;
$$;

CREATE OR REPLACE FUNCTION list_inventories()
RETURNS TABLE(
    id UUID, product_variant_id UUID, product_price_id UUID, product_detail_id UUID,
    product_id UUID, name VARCHAR, variant_name VARCHAR, status VARCHAR, total_quantity BIGINT,
    sold_quantity BIGINT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT x.*
    FROM inventories i
    CROSS JOIN LATERAL get_inventory_by_id(i.id) x
    ORDER BY x.created_at DESC;
$$;

CREATE FUNCTION get_product_price_by_id(p_id UUID)
RETURNS TABLE(
    id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR,
    sort_order INT, product_variant_id UUID,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency,
           pp.sort_order, vi.product_variant_id,
           pp.created_at, pp.updated_at
    FROM product_prices pp
    LEFT JOIN LATERAL (
        SELECT v.id AS product_variant_id
        FROM product_variants v
        LEFT JOIN inventories i ON i.product_variant_id = v.id
        WHERE v.product_price_id = pp.id
        ORDER BY (i.id IS NULL), v.created_at, v.id, i.created_at, i.id
        LIMIT 1
    ) vi ON TRUE
    WHERE pp.id = p_id;
$$;

CREATE FUNCTION list_product_prices_by_detail(p_detail_id UUID)
RETURNS TABLE(
    id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR,
    sort_order INT, product_variant_id UUID,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency,
           pp.sort_order, vi.product_variant_id,
           pp.created_at, pp.updated_at
    FROM product_prices pp
    LEFT JOIN LATERAL (
        SELECT v.id AS product_variant_id
        FROM product_variants v
        LEFT JOIN inventories i ON i.product_variant_id = v.id
        WHERE v.product_price_id = pp.id
        ORDER BY (i.id IS NULL), v.created_at, v.id, i.created_at, i.id
        LIMIT 1
    ) vi ON TRUE
    WHERE pp.product_detail_id = p_detail_id
    ORDER BY pp.sort_order, pp.created_at, pp.id;
$$;
