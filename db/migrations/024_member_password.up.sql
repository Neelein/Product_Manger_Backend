CREATE OR REPLACE FUNCTION update_member_password(p_id UUID, p_password VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE members
    SET password = p_password, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;
