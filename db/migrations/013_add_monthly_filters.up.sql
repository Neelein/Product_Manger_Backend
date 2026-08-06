CREATE OR REPLACE FUNCTION list_announcements_by_month(p_year INT, p_month INT, p_limit INT, p_offset INT)
RETURNS TABLE(id UUID, title VARCHAR, content TEXT, image_path VARCHAR, publisher_id UUID, publisher_name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT a.id, a.title, a.content, a.image_path, a.publisher_id, m.name AS publisher_name, a.created_at, a.updated_at
    FROM announcements a
    JOIN members m ON m.id = a.publisher_id
    WHERE DATE_TRUNC('month', a.created_at) = DATE_TRUNC('month', MAKE_DATE(p_year, p_month, 1))
    ORDER BY a.created_at DESC
    LIMIT p_limit
    OFFSET p_offset;
$$;

CREATE OR REPLACE FUNCTION count_announcements_by_month(p_year INT, p_month INT)
RETURNS BIGINT
LANGUAGE sql AS $$
    SELECT COUNT(*)
    FROM announcements
    WHERE DATE_TRUNC('month', created_at) = DATE_TRUNC('month', MAKE_DATE(p_year, p_month, 1));
$$;

CREATE OR REPLACE FUNCTION list_chat_rooms_by_member_by_month(p_member_id UUID, p_year INT, p_month INT)
RETURNS TABLE(
    id UUID, name VARCHAR, created_by UUID,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ,
    last_message_content TEXT, last_message_sender_id UUID, last_message_created_at TIMESTAMPTZ,
    unread_count BIGINT
)
LANGUAGE sql AS $$
    SELECT
        r.id, r.name, r.created_by,
        r.created_at, r.updated_at,
        lm.content AS last_message_content,
        lm.sender_id AS last_message_sender_id,
        lm.created_at AS last_message_created_at,
        COUNT(*) FILTER (WHERE rc.message_id IS NULL) AS unread_count
    FROM chat_rooms r
    JOIN chat_room_members crm ON crm.room_id = r.id AND crm.member_id = p_member_id
    LEFT JOIN LATERAL (
        SELECT m.content, m.sender_id, m.created_at
        FROM chat_messages m
        WHERE m.room_id = r.id
        ORDER BY m.created_at DESC
        LIMIT 1
    ) lm ON TRUE
    LEFT JOIN chat_messages m ON m.room_id = r.id
    LEFT JOIN read_receipts rc ON rc.message_id = m.id AND rc.member_id = p_member_id
    WHERE DATE_TRUNC('month', r.created_at) = DATE_TRUNC('month', MAKE_DATE(p_year, p_month, 1))
    GROUP BY r.id, r.name, r.created_by, r.created_at, r.updated_at, lm.content, lm.sender_id, lm.created_at
    ORDER BY COALESCE(lm.created_at, r.created_at) DESC;
$$;
