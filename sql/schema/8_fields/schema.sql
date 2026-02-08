CREATE TABLE IF NOT EXISTS fields(
    field_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fields_parent ON fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_fields_author ON fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_fields_modified
    AFTER UPDATE ON fields
    FOR EACH ROW
    BEGIN
        UPDATE fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE field_id = NEW.field_id;
    END;
