CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL UNIQUE,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE
);
