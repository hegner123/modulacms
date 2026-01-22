CREATE TABLE admin_datatypes (
    admin_datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_author ON admin_datatypes(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_datatypes_modified
    AFTER UPDATE ON admin_datatypes
    FOR EACH ROW
    BEGIN
        UPDATE admin_datatypes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_datatype_id = NEW.admin_datatype_id;
    END;
