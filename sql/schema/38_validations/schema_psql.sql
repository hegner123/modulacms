CREATE TABLE IF NOT EXISTS validations (
    validation_id TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    config        TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_validations_name ON validations(name);
CREATE INDEX IF NOT EXISTS idx_validations_author ON validations(author_id);

CREATE OR REPLACE FUNCTION update_validations_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_validations_modified_trigger
    BEFORE UPDATE ON validations
    FOR EACH ROW
    EXECUTE FUNCTION update_validations_modified();
