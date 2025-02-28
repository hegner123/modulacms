CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL unique,
    permissions TEXT NOT NULL unique
);
