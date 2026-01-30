CREATE TABLE IF NOT EXISTS datatypes(
    datatype_id TEXT PRIMARY KEY NOT NULL CHECK (length(datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_author ON datatypes(author_id);

CREATE TRIGGER IF NOT EXISTS update_datatypes_modified
    AFTER UPDATE ON datatypes
    FOR EACH ROW
    BEGIN
        UPDATE datatypes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE datatype_id = NEW.datatype_id;
    END;
