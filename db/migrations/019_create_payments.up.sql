CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE RESTRICT,
    method VARCHAR(30) NOT NULL CHECK (method = 'credit_card'),
    status VARCHAR(20) NOT NULL CHECK (status IN ('paid','failed')),
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    masked_card VARCHAR(32) NOT NULL,
    last4 CHAR(4) NOT NULL CHECK (last4 ~ '^[0-9]{4}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX payments_one_per_order ON payments(order_id);
CREATE INDEX payments_member_id_idx ON payments(member_id);
CREATE INDEX orders_pending_expiration_idx ON orders(created_at) WHERE status='pending' AND payment_status='pending';
