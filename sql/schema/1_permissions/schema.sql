CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL CHECK (length(permission_id) = 26),
    table_id TEXT NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);
