CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id SERIAL PRIMARY KEY,
    admin_route_id INTEGER,
    admin_datatype_id INTEGER,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    CONSTRAINT fk_admin_routes FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);
