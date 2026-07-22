CREATE OR REPLACE FUNCTION list_members_not_in_room(p_room_id UUID, p_limit INT, p_offset INT)
RETURNS TABLE(id UUID, name VARCHAR)
LANGUAGE sql AS $$
    SELECT m.id, m.name
    FROM members m
    WHERE NOT EXISTS (
        SELECT 1 FROM chat_room_members crm
        WHERE crm.room_id = p_room_id AND crm.member_id = m.id
    )
    ORDER BY m.name ASC
    LIMIT p_limit OFFSET p_offset;
$$;

CREATE OR REPLACE FUNCTION count_members_not_in_room(p_room_id UUID)
RETURNS BIGINT
LANGUAGE sql AS $$
    SELECT COUNT(*) FROM members m
    WHERE NOT EXISTS (
        SELECT 1 FROM chat_room_members crm
        WHERE crm.room_id = p_room_id AND crm.member_id = m.id
    );
$$;
