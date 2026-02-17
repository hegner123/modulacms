CREATE TABLE admin_fields (
    admin_field_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
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
