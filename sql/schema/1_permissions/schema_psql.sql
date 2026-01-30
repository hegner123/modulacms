CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL,
    table_id TEXT NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);
