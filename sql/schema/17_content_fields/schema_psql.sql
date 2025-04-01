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
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
