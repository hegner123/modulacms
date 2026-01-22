CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id SERIAL
        PRIMARY KEY,
    route_id INTEGER
        CONSTRAINT fk_route_id
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        CONSTRAINT fk_content_data
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_fields_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_fields_route ON content_fields(route_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_field ON content_fields(field_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_author ON content_fields(author_id);

CREATE OR REPLACE FUNCTION update_content_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_content_fields_modified_trigger
    BEFORE UPDATE ON content_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_content_fields_modified();
