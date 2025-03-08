CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL PRIMARY KEY,
    route_id INTEGER,
    datatype_id INTEGER,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

