-- ============================================================
-- Migration 015: member role + registration codes
-- Registration now requires a one-time-use registration code.
-- ============================================================

ALTER TABLE members
    ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'member';

-- Promote the seeded admin account created in migration 014.
UPDATE members
SET role = 'admin'
WHERE email = 'shakya1221@gmail.com';

CREATE TABLE registration_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(64) NOT NULL UNIQUE,
    created_by  UUID REFERENCES members(id) ON DELETE SET NULL,
    used_by     UUID REFERENCES members(id) ON DELETE RESTRICT,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_registration_codes_created_by ON registration_codes(created_by);
CREATE INDEX idx_registration_codes_used_by ON registration_codes(used_by);

-- Existing member lookups must also return the role.
-- (DROP + CREATE: the return type changes from 6 to 7 columns, which
-- CREATE OR REPLACE FUNCTION cannot do.)
DROP FUNCTION IF EXISTS get_member_by_email(VARCHAR) CASCADE;
CREATE FUNCTION get_member_by_email(p_email VARCHAR)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, role VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, role, created_at, updated_at
    FROM members
    WHERE email = p_email;
$$;

DROP FUNCTION IF EXISTS get_member_by_id(UUID) CASCADE;
CREATE FUNCTION get_member_by_id(p_id UUID)
RETURNS TABLE(id UUID, email VARCHAR, password VARCHAR, name VARCHAR, role VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT id, email, password, name, role, created_at, updated_at
    FROM members
    WHERE id = p_id;
$$;

-- Registers a member only if the registration code exists and is unused.
-- Atomically consumes the code inside the same transaction so a code can
-- never be used twice, even under concurrent registration.
CREATE OR REPLACE FUNCTION register_member_with_code(p_email VARCHAR, p_password VARCHAR, p_name VARCHAR, p_code VARCHAR)
RETURNS TABLE(id UUID, role VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
DECLARE
    v_code_id UUID;
    v_member_id UUID;
    v_created_at TIMESTAMPTZ;
    v_updated_at TIMESTAMPTZ;
BEGIN
    -- Claim an unused code (FOR UPDATE serializes concurrent attempts).
    SELECT rc.id INTO v_code_id
    FROM registration_codes rc
    WHERE rc.code = p_code AND rc.used_by IS NULL
    FOR UPDATE;

    IF NOT FOUND THEN
        IF EXISTS (SELECT 1 FROM registration_codes WHERE code = p_code) THEN
            RAISE EXCEPTION 'registration code already used' USING ERRCODE = 'R0002';
        END IF;
        RAISE EXCEPTION 'registration code does not exist' USING ERRCODE = 'R0001';
    END IF;

    INSERT INTO members (email, password, name)
    VALUES (p_email, p_password, p_name)
    RETURNING members.id, members.created_at, members.updated_at INTO v_member_id, v_created_at, v_updated_at;

    UPDATE registration_codes
    SET used_by = v_member_id, used_at = now()
    WHERE registration_codes.id = v_code_id;

    RETURN QUERY SELECT v_member_id, 'member'::VARCHAR, v_created_at, v_updated_at;
END;
$$;

CREATE OR REPLACE FUNCTION create_registration_code(p_created_by UUID, p_code VARCHAR DEFAULT NULL)
RETURNS TABLE(id UUID, code VARCHAR, created_at TIMESTAMPTZ)
LANGUAGE plpgsql AS $$
DECLARE
    v_code VARCHAR := COALESCE(NULLIF(NULLIF(p_code, ''), NULL), upper(substr(md5(random()::text || clock_timestamp()::text), 1, 8)));
BEGIN
    RETURN QUERY
    INSERT INTO registration_codes (code, created_by)
    VALUES (v_code, p_created_by)
    RETURNING registration_codes.id, registration_codes.code, registration_codes.created_at;
END;
$$;

CREATE OR REPLACE FUNCTION list_registration_codes()
RETURNS TABLE(id UUID, code VARCHAR, created_by TEXT, created_by_email TEXT, used_by TEXT, used_by_email TEXT, used_at TIMESTAMPTZ, created_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT
        rc.id,
        rc.code,
        COALESCE(rc.created_by::text, ''),
        COALESCE(cb.email, ''),
        COALESCE(rc.used_by::text, ''),
        COALESCE(ub.email, ''),
        rc.used_at,
        rc.created_at
    FROM registration_codes rc
    LEFT JOIN members cb ON cb.id = rc.created_by
    LEFT JOIN members ub ON ub.id = rc.used_by
    ORDER BY rc.created_at DESC;
$$;

CREATE OR REPLACE FUNCTION delete_registration_code(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
DECLARE
    v_deleted BOOLEAN := FALSE;
BEGIN
    DELETE FROM registration_codes WHERE id = p_id AND used_at IS NULL
    RETURNING TRUE INTO v_deleted;
    IF v_deleted IS NULL THEN
        v_deleted := FALSE;
    END IF;
    RETURN v_deleted;
END;
$$;
