-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id SERIAL PRIMARY KEY,
    admin_route_id INTEGER,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_data FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);
