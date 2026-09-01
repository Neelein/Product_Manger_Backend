CREATE FUNCTION delete_product_image(p_product_id UUID, p_image_id UUID)
RETURNS TABLE(id UUID, product_id UUID, filename VARCHAR, created_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    DELETE FROM product_images
    WHERE product_images.product_id = p_product_id
      AND product_images.id = p_image_id
    RETURNING product_images.id, product_images.product_id, product_images.filename, product_images.created_at;
$$;
