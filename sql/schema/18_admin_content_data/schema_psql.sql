CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_route_id TEXT NOT NULL
        CONSTRAINT fk_admin_routes
            REFERENCES admin_routes
            ON UPDATE CASCADE,
    admin_datatype_id TEXT NOT NULL
        CONSTRAINT fk_admin_datatypes
            REFERENCES admin_datatypes
            ON UPDATE CASCADE,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_author ON admin_content_data(author_id);

CREATE OR REPLACE FUNCTION update_admin_content_data_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_content_data_modified_trigger
    BEFORE UPDATE ON admin_content_data
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_content_data_modified();
