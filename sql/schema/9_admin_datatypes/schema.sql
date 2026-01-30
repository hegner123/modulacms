CREATE TABLE admin_datatypes (
    admin_datatype_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
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
