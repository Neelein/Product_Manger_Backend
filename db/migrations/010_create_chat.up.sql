CREATE TABLE chat_rooms (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255),
    created_by  UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chat_messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id   UUID NOT NULL REFERENCES members(id),
    content     TEXT NOT NULL DEFAULT '',
    image_path  VARCHAR(500) NOT NULL DEFAULT '',
    file_path   VARCHAR(500) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chat_room_members (
    room_id     UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, member_id)
);

CREATE TABLE read_receipts (
    message_id  UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    member_id   UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    read_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, member_id)
);

CREATE INDEX idx_chat_messages_room_id_created_at ON chat_messages (room_id, created_at DESC);
CREATE INDEX idx_chat_room_members_member_id ON chat_room_members (member_id);
CREATE INDEX idx_read_receipts_message_id ON read_receipts (message_id);

-- ========== CHAT ROOM FUNCTIONS ==========

CREATE OR REPLACE FUNCTION create_chat_room(
    p_name VARCHAR,
    p_created_by UUID
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO chat_rooms (name, created_by)
    VALUES (p_name, p_created_by)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION add_chat_room_members(
    p_room_id UUID,
    p_member_ids UUID[]
)
RETURNS TABLE(member_id UUID)
LANGUAGE sql AS $$
    INSERT INTO chat_room_members (room_id, member_id)
    SELECT p_room_id, unnest(p_member_ids)
    ON CONFLICT DO NOTHING
    RETURNING member_id;
$$;

CREATE OR REPLACE FUNCTION get_chat_room_by_id(
    p_room_id UUID,
    p_member_id UUID
)
RETURNS TABLE(
    id UUID, name VARCHAR, created_by UUID,
    created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ,
    is_member BOOLEAN
)
LANGUAGE sql AS $$
    SELECT
        r.id, r.name, r.created_by,
        r.created_at, r.updated_at,
        EXISTS(SELECT 1 FROM chat_room_members m WHERE m.room_id = r.id AND m.member_id = p_member_id) AS is_member
    FROM chat_rooms r
    WHERE r.id = p_room_id;
$$;

CREATE OR REPLACE FUNCTION list_chat_rooms_by_member(p_member_id UUID)
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
    GROUP BY r.id, r.name, r.created_by, r.created_at, r.updated_at, lm.content, lm.sender_id, lm.created_at
    ORDER BY COALESCE(lm.created_at, r.created_at) DESC;
$$;

CREATE OR REPLACE FUNCTION update_chat_room(p_id UUID, p_name VARCHAR)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE chat_rooms
    SET name = p_name, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_chat_room(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM chat_rooms WHERE id = p_id;
    RETURN FOUND;
END;
$$;

CREATE OR REPLACE FUNCTION remove_chat_room_member(p_room_id UUID, p_member_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM chat_room_members WHERE room_id = p_room_id AND member_id = p_member_id;
    RETURN FOUND;
END;
$$;

-- ========== MESSAGE FUNCTIONS ==========

CREATE OR REPLACE FUNCTION send_message(
    p_room_id UUID, p_sender_id UUID, p_content TEXT,
    p_image_path VARCHAR, p_file_path VARCHAR
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO chat_messages (room_id, sender_id, content, image_path, file_path)
    VALUES (p_room_id, p_sender_id, p_content, p_image_path, p_file_path)
    RETURNING id, created_at;
$$;

CREATE OR REPLACE FUNCTION list_messages(
    p_room_id UUID, p_before_id UUID, p_limit INT
)
RETURNS TABLE(
    id UUID, room_id UUID, sender_id UUID, content TEXT,
    image_path VARCHAR, file_path VARCHAR, created_at TIMESTAMPTZ,
    sender_name VARCHAR
)
LANGUAGE sql AS $$
    SELECT
        m.id, m.room_id, m.sender_id, m.content,
        m.image_path, m.file_path, m.created_at,
        mb.name AS sender_name
    FROM chat_messages m
    JOIN members mb ON mb.id = m.sender_id
    WHERE m.room_id = p_room_id
      AND (p_before_id IS NULL OR m.created_at < (SELECT created_at FROM chat_messages WHERE id = p_before_id))
    ORDER BY m.created_at DESC
    LIMIT p_limit;
$$;

CREATE OR REPLACE FUNCTION delete_message(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM chat_messages WHERE id = p_id;
    RETURN FOUND;
END;
$$;

-- ========== READ RECEIPT FUNCTIONS ==========

CREATE OR REPLACE FUNCTION mark_message_read(p_message_id UUID, p_member_id UUID)
RETURNS TABLE(read_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO read_receipts (message_id, member_id)
    VALUES (p_message_id, p_member_id)
    ON CONFLICT (message_id, member_id) DO UPDATE SET read_at = now()
    RETURNING read_at;
$$;

CREATE OR REPLACE FUNCTION get_message_read_by(p_message_id UUID)
RETURNS TABLE(member_id UUID, member_name VARCHAR, read_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT rc.member_id, m.name AS member_name, rc.read_at
    FROM read_receipts rc
    JOIN members m ON m.id = rc.member_id
    WHERE rc.message_id = p_message_id
    ORDER BY rc.read_at;
$$;

CREATE OR REPLACE FUNCTION count_room_unread(p_room_id UUID, p_member_id UUID)
RETURNS BIGINT
LANGUAGE sql AS $$
    SELECT COUNT(*)
    FROM chat_messages m
    WHERE m.room_id = p_room_id
      AND m.sender_id != p_member_id
      AND NOT EXISTS (
          SELECT 1 FROM read_receipts rc
          WHERE rc.message_id = m.id AND rc.member_id = p_member_id
      );
$$;
