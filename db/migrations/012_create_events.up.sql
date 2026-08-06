CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    status      VARCHAR(50) NOT NULL DEFAULT 'active',
    created_by  UUID NOT NULL REFERENCES members(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_viewers (
    event_id   UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    member_id  UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, member_id)
);

CREATE OR REPLACE FUNCTION create_event(
    p_title VARCHAR, p_description TEXT, p_start_time TIMESTAMPTZ, p_end_time TIMESTAMPTZ, p_status VARCHAR, p_created_by UUID
)
RETURNS TABLE(id UUID, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO events (title, description, start_time, end_time, status, created_by)
    VALUES (p_title, p_description, p_start_time, p_end_time, p_status, p_created_by)
    RETURNING id, created_at, updated_at;
$$;

CREATE OR REPLACE FUNCTION get_event_by_id(p_id UUID, p_member_id UUID)
RETURNS TABLE(id UUID, title VARCHAR, description TEXT, start_time TIMESTAMPTZ, end_time TIMESTAMPTZ, status VARCHAR, created_by UUID, creator_name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT e.id, e.title, e.description, e.start_time, e.end_time, e.status, e.created_by, m.name AS creator_name, e.created_at, e.updated_at
    FROM events e
    JOIN members m ON m.id = e.created_by
    WHERE e.id = p_id
      AND (e.created_by = p_member_id OR e.id IN (SELECT event_id FROM event_viewers WHERE member_id = p_member_id));
$$;

CREATE OR REPLACE FUNCTION list_events_by_month(p_year INT, p_month INT, p_member_id UUID)
RETURNS TABLE(id UUID, title VARCHAR, description TEXT, start_time TIMESTAMPTZ, end_time TIMESTAMPTZ, status VARCHAR, created_by UUID, creator_name VARCHAR, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    SELECT e.id, e.title, e.description, e.start_time, e.end_time, e.status, e.created_by, m.name AS creator_name, e.created_at, e.updated_at
    FROM events e
    JOIN members m ON m.id = e.created_by
    WHERE DATE_TRUNC('month', e.start_time) = DATE_TRUNC('month', MAKE_DATE(p_year, p_month, 1))
      AND (e.created_by = p_member_id OR e.id IN (SELECT event_id FROM event_viewers WHERE member_id = p_member_id))
    ORDER BY e.start_time ASC;
$$;

CREATE OR REPLACE FUNCTION update_event(
    p_id UUID, p_title VARCHAR, p_description TEXT, p_start_time TIMESTAMPTZ, p_end_time TIMESTAMPTZ, p_status VARCHAR
)
RETURNS TABLE(updated_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    UPDATE events
    SET title = p_title, description = p_description, start_time = p_start_time, end_time = p_end_time, status = p_status, updated_at = now()
    WHERE id = p_id
    RETURNING updated_at;
$$;

CREATE OR REPLACE FUNCTION delete_event(p_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM events WHERE id = p_id;
    RETURN FOUND;
END;
$$;

CREATE OR REPLACE FUNCTION add_event_viewer(p_event_id UUID, p_member_id UUID)
RETURNS TABLE(created_at TIMESTAMPTZ)
LANGUAGE sql AS $$
    INSERT INTO event_viewers (event_id, member_id)
    VALUES (p_event_id, p_member_id)
    ON CONFLICT DO NOTHING
    RETURNING created_at;
$$;

CREATE OR REPLACE FUNCTION remove_event_viewer(p_event_id UUID, p_member_id UUID)
RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM event_viewers WHERE event_id = p_event_id AND member_id = p_member_id;
    RETURN FOUND;
END;
$$;

CREATE OR REPLACE FUNCTION list_event_viewers(p_event_id UUID)
RETURNS TABLE(member_id UUID, name VARCHAR)
LANGUAGE sql AS $$
    SELECT ev.member_id, m.name
    FROM event_viewers ev
    JOIN members m ON m.id = ev.member_id
    WHERE ev.event_id = p_event_id;
$$;
