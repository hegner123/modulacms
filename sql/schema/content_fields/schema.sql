CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    content_data_id       INTEGER NOT NULL
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_value         TEXT NOT NULL,
    history TEXT,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP
);

