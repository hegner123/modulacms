CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER
        PRIMARY KEY,
    route_id INTEGER
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_fields_route ON content_fields(route_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_field ON content_fields(field_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_author ON content_fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_content_fields_modified
    AFTER UPDATE ON content_fields
    FOR EACH ROW
    BEGIN
        UPDATE content_fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE content_field_id = NEW.content_field_id;
    END;
