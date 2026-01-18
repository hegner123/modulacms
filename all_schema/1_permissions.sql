CREATE TABLE IF NOT EXISTS permissions (
    permission_id INTEGER
        PRIMARY KEY,
    table_id INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);
