CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);
