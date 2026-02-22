CREATE TABLE IF NOT EXISTS field_types (
    field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);
