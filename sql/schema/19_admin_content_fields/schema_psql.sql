CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id SERIAL
        PRIMARY KEY,
    admin_route_id INTEGER
        CONSTRAINT fk_admin_route_id
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id INTEGER NOT NULL
        CONSTRAINT fk_admin_content_data
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_admin_fields
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
