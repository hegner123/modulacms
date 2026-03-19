CREATE TABLE IF NOT EXISTS fields(
    field_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation_id TEXT DEFAULT NULL
        REFERENCES validations(validation_id)
            ON DELETE SET NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    translatable INTEGER NOT NULL DEFAULT 0,
    roles TEXT DEFAULT NULL,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fields_parent ON fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_fields_author ON fields(author_id);
CREATE INDEX IF NOT EXISTS idx_fields_validation ON fields(validation_id);

CREATE TRIGGER IF NOT EXISTS update_fields_modified
    AFTER UPDATE ON fields
    FOR EACH ROW
    BEGIN
        UPDATE fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE field_id = NEW.field_id;
    END;
