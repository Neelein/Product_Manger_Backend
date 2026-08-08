-- ============================================================
-- Migration 016: categories table
-- products.category (free text) is replaced by products.category_id
-- referencing categories. Product read functions LEFT JOIN the name back.
-- ============================================================

CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Backfill legacy distinct non-empty products.category values.
INSERT INTO categories (name)
SELECT DISTINCT category
FROM products
WHERE category IS NOT NULL AND category <> ''
ON CONFLICT (name) DO NOTHING;

ALTER TABLE products
ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE RESTRICT;

-- Point existing products at their backfilled category, if any.
UPDATE products p
SET category_id = c.id
FROM categories c
WHERE p.category_id IS NULL AND c.name = p.category;

CREATE INDEX idx_products_category_id ON products(category_id);

ALTER TABLE products DROP COLUMN category;

-- Product read/create/update functions: the return shape changed to carry
-- category_id, and create/update now take a category_id.
-- (DROP + CREATE: CREATE OR REPLACE FUNCTION cannot change the RETURNS TABLE.)
DROP FUNCTION IF EXISTS create_product(VARCHAR, VARCHAR, VARCHAR, UUID) CASCADE;
DROP FUNCTION IF EXISTS update_product(UUID, VARCHAR, VARCHAR, VARCHAR) CASCADE;
DROP FUNCTION IF EXISTS list_products() CASCADE;
DROP FUNCTION IF EXISTS get_product_by_id(UUID) CASCADE;

CREATE FUNCTION create_product(p_name VARCHAR, p_status VARCHAR, p_category_id UUID, p_member_id UUID)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO products (type, name, status, category_id, member_id)
    VALUES ('product', p_name, p_status, p_category_id, p_member_id)
    RETURNING id, created_at, updated_at;
$$;

CREATE FUNCTION update_product(p_id UUID, p_name VARCHAR, p_status VARCHAR, p_category_id UUID)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE products
    SET name = p_name, status = p_status, category_id = p_category_id, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE FUNCTION list_products()
RETURNS TABLE(id UUID, name VARCHAR, status VARCHAR, category_id VARCHAR, category VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT p.id, p.name, p.status, COALESCE(p.category_id::text, ''), COALESCE(c.name, ''), p.created_at, p.updated_at
    FROM products p
    LEFT JOIN categories c ON c.id = p.category_id
    ORDER BY p.created_at DESC;
$$;

CREATE FUNCTION get_product_by_id(p_id UUID)
RETURNS TABLE(id UUID, name VARCHAR, status VARCHAR, category_id VARCHAR, category VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT p.id, p.name, p.status, COALESCE(p.category_id::text, ''), COALESCE(c.name, ''), p.created_at, p.updated_at
    FROM products p
    LEFT JOIN categories c ON c.id = p.category_id
    WHERE p.id = p_id;
$$;

-- ------------------------------------------------------------
-- Categories
-- ------------------------------------------------------------

CREATE FUNCTION create_category(p_name VARCHAR)
RETURNS TABLE(id UUID, name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO categories (name)
    VALUES (p_name)
    RETURNING id, name, created_at, updated_at;
$$;

CREATE FUNCTION list_categories()
RETURNS TABLE(id UUID, name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, name, created_at, updated_at
    FROM categories
    ORDER BY created_at DESC;
$$;

CREATE FUNCTION update_category(p_id UUID, p_name VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE categories
    SET name = p_name, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE FUNCTION delete_category(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
DECLARE
    v_deleted BOOLEAN := FALSE;
BEGIN
    IF EXISTS (SELECT 1 FROM products WHERE products.category_id = p_id) THEN
        RAISE EXCEPTION 'category is in use by products' USING ERRCODE = 'R0010';
    END IF;

    DELETE FROM categories WHERE categories.id = p_id
    RETURNING TRUE INTO v_deleted;
    IF v_deleted IS NULL THEN
        v_deleted := FALSE;
    END IF;

    RETURN v_deleted;
END;
$$;