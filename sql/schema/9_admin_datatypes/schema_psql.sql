CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE RESTRICT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_author ON admin_datatypes(author_id);

CREATE OR REPLACE FUNCTION update_admin_datatypes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_datatypes_modified_trigger
    BEFORE UPDATE ON admin_datatypes
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_datatypes_modified();
