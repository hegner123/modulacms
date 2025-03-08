CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id      INTEGER NOT NULL
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_datatype_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
