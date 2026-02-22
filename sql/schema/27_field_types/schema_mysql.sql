CREATE TABLE IF NOT EXISTS field_types (
    field_type_id VARCHAR(26) PRIMARY KEY NOT NULL,
    type VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    CONSTRAINT field_types_type_unique UNIQUE (type)
);
