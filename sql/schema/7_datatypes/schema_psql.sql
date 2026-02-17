CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_author ON datatypes(author_id);

CREATE OR REPLACE FUNCTION update_datatypes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_datatypes_modified_trigger
    BEFORE UPDATE ON datatypes
    FOR EACH ROW
    EXECUTE FUNCTION update_datatypes_modified();
