CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL CHECK (length(role_id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    permissions TEXT NOT NULL
        UNIQUE
);
