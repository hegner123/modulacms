-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id       INTEGER NOT NULL
    REFERENCES admin_routes(admin_route_id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id       INTEGER NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value         TEXT NOT NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP,
    history             TEXT
);
