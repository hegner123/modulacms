CREATE TABLE content_fields (
    content_field_id INTEGER
        PRIMARY KEY,
    route_id INTEGER NOT NULL
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
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

CREATE INDEX idx_content_fields_content_data_id
    ON content_fields (content_data_id);

CREATE INDEX idx_content_fields_field_id
    ON content_fields (field_id);

CREATE INDEX idx_content_fields_route_id
    ON content_fields (route_id);

