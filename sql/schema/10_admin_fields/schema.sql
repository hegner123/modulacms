CREATE TABLE admin_fields (
    admin_field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_fields_author ON admin_fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_fields_modified
    AFTER UPDATE ON admin_fields
    FOR EACH ROW
    BEGIN
        UPDATE admin_fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_field_id = NEW.admin_field_id;
    END;
