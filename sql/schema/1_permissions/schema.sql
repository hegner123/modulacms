CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL CHECK (length(permission_id) = 26),
    label TEXT NOT NULL UNIQUE,
    system_protected INTEGER NOT NULL DEFAULT 0
);
