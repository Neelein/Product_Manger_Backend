DROP FUNCTION IF EXISTS get_product_price_by_id(uuid);
DROP FUNCTION IF EXISTS list_product_prices_by_detail(uuid);

CREATE FUNCTION get_product_price_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency, pp.sort_order, pp.created_at, pp.updated_at
    FROM product_prices pp
    WHERE pp.id = p_id;
$$;

CREATE FUNCTION list_product_prices_by_detail(p_product_detail_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT pp.id, pp.product_detail_id, pp.label, pp.amount, pp.currency, pp.sort_order, pp.created_at, pp.updated_at
    FROM product_prices pp
    WHERE pp.product_detail_id = p_product_detail_id
    ORDER BY pp.sort_order;
$$;
