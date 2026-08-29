-- Member types and unconstrained permissions.
ALTER TABLE members
    ADD COLUMN member_type VARCHAR(20) NOT NULL DEFAULT 'customer',
    ADD COLUMN permission VARCHAR(100);

-- Only explicit administrative or employee roles cross the employee boundary.
-- Legacy member rows remain customers with no permission.
UPDATE members
SET member_type = CASE WHEN lower(role) IN ('admin', 'employee') THEN 'employee' ELSE 'customer' END,
    permission = CASE
        WHEN lower(role) = 'admin' THEN 'admin'
        WHEN lower(role) = 'employee' THEN 'employee'
        ELSE NULL
    END
WHERE role IS NOT NULL;

ALTER TABLE members
    DROP COLUMN role;

ALTER TABLE members
    ADD CONSTRAINT members_member_type_check CHECK (member_type IN ('customer', 'employee'));

DROP FUNCTION IF EXISTS create_member(VARCHAR, VARCHAR, VARCHAR);
CREATE FUNCTION create_member(p_email VARCHAR, p_password VARCHAR, p_name VARCHAR)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO members (email, password, name, member_type, permission)
    VALUES (p_email, p_password, p_name, 'customer', NULL)
    RETURNING id, created_at, updated_at;
$$;

DROP FUNCTION IF EXISTS get_member_by_email(VARCHAR) CASCADE;
CREATE FUNCTION get_member_by_email(p_email VARCHAR)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, member_type VARCHAR, permission VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, member_type, permission, created_at, updated_at
    FROM members WHERE email = p_email;
$$;

DROP FUNCTION IF EXISTS get_member_by_id(UUID) CASCADE;
CREATE FUNCTION get_member_by_id(p_id UUID)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, member_type VARCHAR, permission VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, member_type, permission, created_at, updated_at
    FROM members WHERE id = p_id;
$$;

DROP FUNCTION IF EXISTS register_member_with_code(VARCHAR, VARCHAR, VARCHAR, VARCHAR);
CREATE FUNCTION register_member_with_code(p_email VARCHAR, p_password VARCHAR, p_name VARCHAR, p_code VARCHAR)
RETURNS TABLE(id UUID, member_type VARCHAR, permission VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
DECLARE v_code_id UUID; v_member_id UUID; v_created_at TIMESTAMPTZ; v_updated_at TIMESTAMPTZ;
BEGIN
    SELECT rc.id INTO v_code_id FROM registration_codes AS rc WHERE rc.code = p_code AND rc.used_by IS NULL FOR UPDATE;
    IF NOT FOUND THEN
        IF EXISTS (SELECT 1 FROM registration_codes WHERE code = p_code) THEN
            RAISE EXCEPTION 'registration code already used' USING ERRCODE = 'R0002';
        END IF;
        RAISE EXCEPTION 'registration code does not exist' USING ERRCODE = 'R0001';
    END IF;
    INSERT INTO members (email, password, name, member_type, permission)
    VALUES (p_email, p_password, p_name, 'employee', NULL)
    RETURNING members.id, members.created_at, members.updated_at INTO v_member_id, v_created_at, v_updated_at;
    UPDATE registration_codes AS rc SET used_by = v_member_id, used_at = now() WHERE rc.id = v_code_id;
    RETURN QUERY SELECT v_member_id, 'employee'::VARCHAR, NULL::VARCHAR, v_created_at, v_updated_at;
END;
$$;

CREATE OR REPLACE FUNCTION update_member_permission(p_id UUID, p_permission VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE members SET permission = p_permission, updated_at = now()
    WHERE id = p_id AND member_type = 'employee'
    RETURNING updated_at;
$$;

ALTER TABLE registration_codes
    DROP CONSTRAINT IF EXISTS registration_codes_used_by_fkey,
    ADD CONSTRAINT registration_codes_used_by_fkey
        FOREIGN KEY (used_by) REFERENCES members(id) ON DELETE RESTRICT;

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_no VARCHAR(64),
    customer_id UUID NOT NULL REFERENCES members(id),
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','confirmed','cancelled','completed')),
    payment_status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (payment_status IN ('pending','paid','failed','refunded')),
    fulfillment_status VARCHAR(30) NOT NULL DEFAULT 'unfulfilled' CHECK (fulfillment_status IN ('unfulfilled','fulfilled','cancelled')),
    subtotal NUMERIC(12,2) NOT NULL CHECK (subtotal >= 0),
    total_amount NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
    customer_snapshot JSONB NOT NULL,
    shipping_address_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

UPDATE orders SET order_no = 'ORD-' || id::text WHERE order_no IS NULL;
ALTER TABLE orders
    ALTER COLUMN order_no SET NOT NULL,
    ADD CONSTRAINT orders_order_no_key UNIQUE (order_no);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    line_total NUMERIC(12,2) NOT NULL CHECK (line_total >= 0),
    product_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE order_items
    ADD CONSTRAINT order_items_id_order_id_key UNIQUE (id, order_id);

CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL,
    changed_by UUID REFERENCES members(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inventory_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    inventory_item_id UUID NOT NULL REFERENCES inventory_items(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','released','fulfilled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    released_at TIMESTAMPTZ
);

ALTER TABLE inventory_reservations
    ADD CONSTRAINT inventory_reservations_order_item_order_fkey
    FOREIGN KEY (order_item_id, order_id)
    REFERENCES order_items (id, order_id) ON DELETE CASCADE;

CREATE UNIQUE INDEX inventory_reservations_one_active_item
    ON inventory_reservations(inventory_item_id) WHERE status = 'active';
CREATE INDEX orders_customer_id_idx ON orders(customer_id);
CREATE INDEX order_items_order_id_idx ON order_items(order_id);
