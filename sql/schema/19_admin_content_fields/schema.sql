CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER
        PRIMARY KEY,
    admin_route_id INTEGER NOT NULL
        REFERENCES admin_routes
            ON DELETE SET NULL,
    admin_content_data_id INTEGER NOT NULL
        REFERENCES admin_content_data
            ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        REFERENCES admin_fields
            ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

