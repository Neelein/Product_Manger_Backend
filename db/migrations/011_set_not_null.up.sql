-- ============================================================
-- Migration 011: Make all nullable columns NOT NULL with defaults
-- ============================================================

-- Create a system member for the default UUID on member_id
INSERT INTO members (id, email, password, name)
VALUES ('00000000-0000-0000-0000-000000000000', 'system@internal', '', 'System')
ON CONFLICT (id) DO NOTHING;

-- products.category
UPDATE products SET category = '' WHERE category IS NULL;
ALTER TABLE products ALTER COLUMN category SET DEFAULT '';
ALTER TABLE products ALTER COLUMN category SET NOT NULL;

-- products.member_id
UPDATE products SET member_id = '00000000-0000-0000-0000-000000000000' WHERE member_id IS NULL;
ALTER TABLE products DROP CONSTRAINT products_member_id_fkey;
ALTER TABLE products ALTER COLUMN member_id SET DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE products ALTER COLUMN member_id SET NOT NULL;
ALTER TABLE products ADD CONSTRAINT products_member_id_fkey
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET DEFAULT;

-- product_details.introduction
UPDATE product_details SET introduction = '' WHERE introduction IS NULL;
ALTER TABLE product_details ALTER COLUMN introduction SET DEFAULT '';
ALTER TABLE product_details ALTER COLUMN introduction SET NOT NULL;

-- product_details.usage_instructions
UPDATE product_details SET usage_instructions = '' WHERE usage_instructions IS NULL;
ALTER TABLE product_details ALTER COLUMN usage_instructions SET DEFAULT '';
ALTER TABLE product_details ALTER COLUMN usage_instructions SET NOT NULL;

-- product_details.return_policy
UPDATE product_details SET return_policy = '' WHERE return_policy IS NULL;
ALTER TABLE product_details ALTER COLUMN return_policy SET DEFAULT '';
ALTER TABLE product_details ALTER COLUMN return_policy SET NOT NULL;

-- inventory_items.cost
UPDATE inventory_items SET cost = 0 WHERE cost IS NULL;
ALTER TABLE inventory_items ALTER COLUMN cost SET DEFAULT 0;
ALTER TABLE inventory_items ALTER COLUMN cost SET NOT NULL;

-- announcements.image_path
UPDATE announcements SET image_path = '' WHERE image_path IS NULL;
ALTER TABLE announcements ALTER COLUMN image_path SET DEFAULT '';
ALTER TABLE announcements ALTER COLUMN image_path SET NOT NULL;
