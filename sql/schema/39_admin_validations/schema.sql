CREATE TABLE IF NOT EXISTS admin_validations (
    admin_validation_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_validation_id) = 26),
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    config              TEXT NOT NULL DEFAULT '{}',
    author_id           TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_validations_name ON admin_validations(name);
CREATE INDEX IF NOT EXISTS idx_admin_validations_author ON admin_validations(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_validations_modified
    AFTER UPDATE ON admin_validations
    FOR EACH ROW
    BEGIN
        UPDATE admin_validations SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_validation_id = NEW.admin_validation_id;
    END;
