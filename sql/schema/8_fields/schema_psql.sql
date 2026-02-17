CREATE TABLE IF NOT EXISTS fields (
    field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id TEXT
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fields_parent ON fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_fields_author ON fields(author_id);

CREATE OR REPLACE FUNCTION update_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_fields_modified_trigger
    BEFORE UPDATE ON fields
    FOR EACH ROW
    EXECUTE FUNCTION update_fields_modified();
