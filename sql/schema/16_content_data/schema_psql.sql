CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    route_id TEXT
        CONSTRAINT fk_routes
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    datatype_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_data_parent ON content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_content_data_route ON content_data(route_id);
CREATE INDEX IF NOT EXISTS idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX IF NOT EXISTS idx_content_data_author ON content_data(author_id);

CREATE OR REPLACE FUNCTION update_content_data_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_content_data_modified_trigger
    BEFORE UPDATE ON content_data
    FOR EACH ROW
    EXECUTE FUNCTION update_content_data_modified();
