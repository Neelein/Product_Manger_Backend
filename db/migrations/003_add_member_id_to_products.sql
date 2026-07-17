ALTER TABLE products
ADD COLUMN member_id UUID REFERENCES members(id) ON DELETE SET NULL;

CREATE INDEX idx_products_member_id ON products(member_id);
