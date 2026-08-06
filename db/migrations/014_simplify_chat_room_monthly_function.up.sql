CREATE OR REPLACE FUNCTION list_chat_rooms_by_member_by_month(p_member_id UUID, p_year INT, p_month INT)
RETURNS TABLE(id UUID, name VARCHAR, created_by UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT r.id, r.name, r.created_by, r.created_at, r.updated_at
    FROM chat_rooms r
    JOIN chat_room_members crm ON crm.room_id = r.id AND crm.member_id = p_member_id
    WHERE DATE_TRUNC('month', r.created_at) = DATE_TRUNC('month', MAKE_DATE(p_year, p_month, 1))
    ORDER BY r.created_at DESC;
$$;
