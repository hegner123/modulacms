CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    permissions JSONB
);

