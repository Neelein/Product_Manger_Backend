CREATE TABLE announcements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        VARCHAR(255) NOT NULL,
    content      TEXT NOT NULL,
    image_path   VARCHAR(500) NOT NULL DEFAULT '',
    publisher_id UUID NOT NULL REFERENCES members(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION create_announcement(
    p_title VARCHAR, p_content TEXT, p_image_path VARCHAR, p_publisher_id UUID
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO announcements (title, content, image_path, publisher_id)
    VALUES (p_title, p_content, p_image_path, p_publisher_id)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_announcement_by_id(p_id UUID)
RETURNS TABLE(id UUID, title VARCHAR, content TEXT, image_path VARCHAR, publisher_id UUID, publisher_name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT a.id, a.title, a.content, a.image_path, a.publisher_id, m.name AS publisher_name, a.created_at, a.updated_at
    FROM announcements a
    JOIN members m ON m.id = a.publisher_id
    WHERE a.id = p_id;
$$;

CREATE OR REPLACE FUNCTION list_announcements(p_limit INT, p_offset INT)
RETURNS TABLE(id UUID, title VARCHAR, content TEXT, image_path VARCHAR, publisher_id UUID, publisher_name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT a.id, a.title, a.content, a.image_path, a.publisher_id, m.name AS publisher_name, a.created_at, a.updated_at
    FROM announcements a
    JOIN members m ON m.id = a.publisher_id
    ORDER BY a.created_at DESC
    LIMIT p_limit
    OFFSET p_offset;
$$;

CREATE OR REPLACE FUNCTION count_announcements()
RETURNS BIGINT
LANGUAGE sql AS $$
    SELECT COUNT(*) FROM announcements;
$$;

CREATE OR REPLACE FUNCTION update_announcement(
    p_id UUID, p_title VARCHAR, p_content TEXT, p_image_path VARCHAR
)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE announcements
    SET title = p_title, content = p_content, image_path = p_image_path, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_announcement(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM announcements WHERE id = p_id;
    RETURN FOUND;
END;
$$;
