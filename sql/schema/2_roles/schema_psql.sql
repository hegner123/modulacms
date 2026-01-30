CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL
        UNIQUE,
    permissions jsonb
);

