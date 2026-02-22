CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id VARCHAR(26) PRIMARY KEY NOT NULL,
    type VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    CONSTRAINT admin_field_types_type_unique UNIQUE (type)
);
