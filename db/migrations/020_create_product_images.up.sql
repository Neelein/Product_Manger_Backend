CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

CREATE FUNCTION create_product_image(p_product_id UUID, p_filename VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(hashtextextended(p_product_id::text, 0));
    IF NOT EXISTS (SELECT 1 FROM products WHERE products.id = p_product_id) THEN
        RAISE EXCEPTION 'product not found' USING ERRCODE = 'P0002';
    END IF;
    IF (SELECT count(*) FROM product_images WHERE product_id = p_product_id) >= 3 THEN
        RAISE EXCEPTION 'product image limit exceeded' USING ERRCODE = 'P0001';
    END IF;
    RETURN QUERY INSERT INTO product_images(product_id, filename)
    VALUES (p_product_id, p_filename)
    RETURNING product_images.id, product_images.created_at;
END;
$$;

CREATE FUNCTION list_product_images(p_product_id UUID)
RETURNS TABLE(id UUID, product_id UUID, filename VARCHAR, created_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, product_id, filename, created_at
    FROM product_images
    WHERE product_id = p_product_id
    ORDER BY created_at ASC;
$$;
