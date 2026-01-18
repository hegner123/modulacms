CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    permissions TEXT NOT NULL
        UNIQUE
);
