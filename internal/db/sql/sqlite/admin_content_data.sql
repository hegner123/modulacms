CREATE TABLE admin_content_data (
    admin_content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER NOT NULL
        REFERENCES admin_content_data
            ON DELETE CASCADE,
    admin_route_id INTEGER NOT NULL
        REFERENCES admin_routes
            ON DELETE CASCADE,
    admin_datatype_id INTEGER NOT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL
);

