CREATE TABLE users (
    id            UUID PRIMARY KEY,
    username      VARCHAR(32) NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          VARCHAR(20) NOT NULL DEFAULT 'user',
    scopes        TEXT[] DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    id            UUID PRIMARY KEY,
    name          VARCHAR(64) NOT NULL,
    host_user_id  UUID NOT NULL REFERENCES users(id),
    max_players   INT NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'waiting',
    players       TEXT[] DEFAULT '{}',
    metadata      JSONB DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_status ON sessions(status);
