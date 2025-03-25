CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL PRIMARY KEY,
    route_id INTEGER,
    parent_id INTEGER,
    datatype_id INTEGER,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);
