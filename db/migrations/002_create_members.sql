CREATE TABLE members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id    UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    session_key  VARCHAR(255) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL DEFAULT now() + interval '1 day'
);

CREATE INDEX idx_sessions_member_id ON sessions(member_id);
CREATE INDEX idx_sessions_session_key ON sessions(session_key);
