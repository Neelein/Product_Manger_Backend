ALTER TABLE product_details
  DROP COLUMN detail_type,
  DROP COLUMN title,
  DROP COLUMN content,
  DROP COLUMN sort_order,
  ADD COLUMN introduction        TEXT NOT NULL DEFAULT '',
  ADD COLUMN usage_instructions  TEXT NOT NULL DEFAULT '',
  ADD COLUMN return_policy       TEXT NOT NULL DEFAULT '';

DROP INDEX IF EXISTS idx_product_details_detail_type;
