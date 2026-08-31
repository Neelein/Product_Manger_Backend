-- Migration 021 introduced variant_name, but databases that already recorded
-- that migration can still contain the older eleven-column function contract.
-- Recreate the inventory read functions so their return definitions and SQL
-- result sets stay aligned with the repository's twelve scan destinations.
DROP FUNCTION IF EXISTS list_inventories();
DROP FUNCTION IF EXISTS get_inventory_by_price_id(UUID);
DROP FUNCTION IF EXISTS get_inventory_by_id(UUID);

CREATE FUNCTION get_inventory_by_id(p_id UUID)
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

CREATE FUNCTION get_inventory_by_price_id(p_product_price_id UUID)
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

CREATE FUNCTION list_inventories()
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
