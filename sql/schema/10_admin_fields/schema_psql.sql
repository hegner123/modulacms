CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT DEFAULT ''::TEXT NOT NULL,
    type TEXT DEFAULT 'text'::TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_fields_author ON admin_fields(author_id);

CREATE OR REPLACE FUNCTION update_admin_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_fields_modified_trigger
    BEFORE UPDATE ON admin_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_fields_modified();
