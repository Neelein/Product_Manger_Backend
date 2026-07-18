-- ============================================================
-- Migration 008: CRUD functions for all domains
-- Go calls SELECT * FROM function_name(...) instead of raw SQL
-- ============================================================

-- ------------------------------------------------------------
-- Members
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_member(p_email VARCHAR, p_password VARCHAR, p_name VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO members (email, password, name)
    VALUES (p_email, p_password, p_name)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_member_by_email(p_email VARCHAR)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, created_at, updated_at
    FROM members
    WHERE email = p_email;
$$;

CREATE OR REPLACE FUNCTION get_member_by_id(p_id UUID)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, created_at, updated_at
    FROM members
    WHERE id = p_id;
$$;

CREATE OR REPLACE FUNCTION update_member(p_id UUID, p_email VARCHAR, p_name VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE members
    SET email = p_email, name = p_name, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

-- ------------------------------------------------------------
-- Products
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_product(
    p_name VARCHAR, p_status VARCHAR, p_category VARCHAR, p_member_id UUID
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO products (type, name, status, category, member_id)
    VALUES ('product', p_name, p_status, p_category, p_member_id)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION list_products()
RETURNS TABLE(id UUID, name VARCHAR, status VARCHAR, category VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, name, status, category, created_at, updated_at
    FROM products
    ORDER BY created_at DESC;
$$;

CREATE OR REPLACE FUNCTION get_product_by_id(p_id UUID)
RETURNS TABLE(id UUID, name VARCHAR, status VARCHAR, category VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, name, status, category, created_at, updated_at
    FROM products
    WHERE id = p_id;
$$;

CREATE OR REPLACE FUNCTION update_product(p_id UUID, p_name VARCHAR, p_status VARCHAR, p_category VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE products
    SET name = p_name, status = p_status, category = p_category, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_product(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM products WHERE id = p_id;
    RETURN FOUND;
END;
$$;

-- ------------------------------------------------------------
-- Product Details
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_product_detail(
    p_product_id UUID, p_introduction TEXT, p_usage_instructions TEXT, p_return_policy TEXT
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO product_details (product_id, introduction, usage_instructions, return_policy)
    VALUES (p_product_id, p_introduction, p_usage_instructions, p_return_policy)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_product_detail_by_product(p_product_id UUID)
RETURNS TABLE(id UUID, product_id UUID, introduction TEXT, usage_instructions TEXT, return_policy TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, product_id, introduction, usage_instructions, return_policy, created_at, updated_at
    FROM product_details
    WHERE product_id = p_product_id;
$$;

CREATE OR REPLACE FUNCTION update_product_detail(
    p_id UUID, p_introduction TEXT, p_usage_instructions TEXT, p_return_policy TEXT
)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE product_details
    SET introduction = p_introduction, usage_instructions = p_usage_instructions, return_policy = p_return_policy, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

-- ------------------------------------------------------------
-- Product Prices
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_product_price(
    p_product_detail_id UUID, p_label VARCHAR, p_amount NUMERIC, p_currency VARCHAR, p_sort_order INT
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO product_prices (product_detail_id, label, amount, currency, sort_order)
    VALUES (p_product_detail_id, p_label, p_amount, COALESCE(p_currency, 'TWD'), p_sort_order)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_product_price_by_id(p_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, product_detail_id, label, amount, currency, sort_order, created_at, updated_at
    FROM product_prices
    WHERE id = p_id;
$$;

CREATE OR REPLACE FUNCTION list_product_prices_by_detail(p_product_detail_id UUID)
RETURNS TABLE(id UUID, product_detail_id UUID, label VARCHAR, amount NUMERIC, currency VARCHAR, sort_order INT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, product_detail_id, label, amount, currency, sort_order, created_at, updated_at
    FROM product_prices
    WHERE product_detail_id = p_product_detail_id
    ORDER BY sort_order;
$$;

CREATE OR REPLACE FUNCTION update_product_price(
    p_id UUID, p_label VARCHAR, p_amount NUMERIC, p_currency VARCHAR, p_sort_order INT
)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE product_prices
    SET label = p_label, amount = p_amount, currency = COALESCE(p_currency, 'TWD'), sort_order = p_sort_order, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

-- ------------------------------------------------------------
-- Inventories (JOIN queries with dynamic computed fields)
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_inventory(p_product_price_id UUID, p_status VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO inventories (product_price_id, status)
    VALUES (p_product_price_id, p_status)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_inventory_by_id(p_id UUID)
RETURNS TABLE(
    id UUID, product_price_id UUID, name VARCHAR, status VARCHAR,
    total_quantity BIGINT, sold_quantity BIGINT,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT i.id, i.product_price_id,
           CONCAT(p.name, '-', pr.label)::VARCHAR AS name,
           i.status,
           COUNT(it.id) AS total_quantity,
           COUNT(it.id) FILTER (WHERE it.status = '出售') AS sold_quantity,
           i.created_at, i.updated_at
    FROM inventories i
    JOIN product_prices pr ON pr.id = i.product_price_id
    JOIN product_details pd ON pd.id = pr.product_detail_id
    JOIN products p ON p.id = pd.product_id
    LEFT JOIN inventory_items it ON it.inventory_id = i.id
    WHERE i.id = p_id
    GROUP BY i.id, i.product_price_id, p.name, pr.label, i.status, i.created_at, i.updated_at;
$$;

CREATE OR REPLACE FUNCTION get_inventory_by_price_id(p_product_price_id UUID)
RETURNS TABLE(
    id UUID, product_price_id UUID, name VARCHAR, status VARCHAR,
    total_quantity BIGINT, sold_quantity BIGINT,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT i.id, i.product_price_id,
           CONCAT(p.name, '-', pr.label)::VARCHAR AS name,
           i.status,
           COUNT(it.id) AS total_quantity,
           COUNT(it.id) FILTER (WHERE it.status = '出售') AS sold_quantity,
           i.created_at, i.updated_at
    FROM inventories i
    JOIN product_prices pr ON pr.id = i.product_price_id
    JOIN product_details pd ON pd.id = pr.product_detail_id
    JOIN products p ON p.id = pd.product_id
    LEFT JOIN inventory_items it ON it.inventory_id = i.id
    WHERE i.product_price_id = p_product_price_id
    GROUP BY i.id, i.product_price_id, p.name, pr.label, i.status, i.created_at, i.updated_at;
$$;

CREATE OR REPLACE FUNCTION list_inventories()
RETURNS TABLE(
    id UUID, product_price_id UUID, name VARCHAR, status VARCHAR,
    total_quantity BIGINT, sold_quantity BIGINT,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
)
LANGUAGE sql AS $$
    SELECT i.id, i.product_price_id,
           CONCAT(p.name, '-', pr.label)::VARCHAR AS name,
           i.status,
           COUNT(it.id) AS total_quantity,
           COUNT(it.id) FILTER (WHERE it.status = '出售') AS sold_quantity,
           i.created_at, i.updated_at
    FROM inventories i
    JOIN product_prices pr ON pr.id = i.product_price_id
    JOIN product_details pd ON pd.id = pr.product_detail_id
    JOIN products p ON p.id = pd.product_id
    LEFT JOIN inventory_items it ON it.inventory_id = i.id
    GROUP BY i.id, i.product_price_id, p.name, pr.label, i.status, i.created_at, i.updated_at
    ORDER BY i.created_at DESC;
$$;

CREATE OR REPLACE FUNCTION update_inventory(p_id UUID, p_status VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE inventories
    SET status = p_status, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_inventory(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM inventories WHERE id = p_id;
    RETURN FOUND;
END;
$$;

-- ------------------------------------------------------------
-- Inventory Items
-- ------------------------------------------------------------

CREATE OR REPLACE FUNCTION create_inventory_item(
    p_inventory_id UUID, p_item_code VARCHAR, p_status VARCHAR,
    p_cost NUMERIC, p_date_added DATE
)
RETURNS TABLE(id UUID, status_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO inventory_items (inventory_id, item_code, status, cost, date_added)
    VALUES (p_inventory_id, p_item_code, p_status, p_cost, p_date_added)
    RETURNING id, status_updated_at, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_inventory_item_by_id(p_id UUID)
RETURNS TABLE(id UUID, inventory_id UUID, item_code VARCHAR, status VARCHAR, cost NUMERIC, date_added DATE, status_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, inventory_id, item_code, status, cost, date_added, status_updated_at, created_at, updated_at
    FROM inventory_items
    WHERE id = p_id;
$$;

CREATE OR REPLACE FUNCTION list_inventory_items(p_inventory_id UUID)
RETURNS TABLE(id UUID, inventory_id UUID, item_code VARCHAR, status VARCHAR, cost NUMERIC, date_added DATE, status_updated_at TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, inventory_id, item_code, status, cost, date_added, status_updated_at, created_at, updated_at
    FROM inventory_items
    WHERE inventory_id = p_inventory_id
    ORDER BY created_at DESC;
$$;

CREATE OR REPLACE FUNCTION update_inventory_item(
    p_id UUID, p_item_code VARCHAR, p_status VARCHAR, p_cost NUMERIC, p_date_added DATE
)
RETURNS TABLE(status_updated_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE inventory_items
    SET item_code = p_item_code, status = p_status, cost = p_cost, date_added = p_date_added,
        status_updated_at = now(), updated_at = now()
    WHERE id = p_id
    RETURNING status_updated_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_inventory_item(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM inventory_items WHERE id = p_id;
    RETURN FOUND;
END;
$$;
