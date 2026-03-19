CREATE TABLE IF NOT EXISTS validations (
    validation_id TEXT PRIMARY KEY NOT NULL CHECK (length(validation_id) = 26),
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    config        TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_validations_name ON validations(name);
CREATE INDEX IF NOT EXISTS idx_validations_author ON validations(author_id);

CREATE TRIGGER IF NOT EXISTS update_validations_modified
    AFTER UPDATE ON validations
    FOR EACH ROW
    BEGIN
        UPDATE validations SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE validation_id = NEW.validation_id;
    END;
