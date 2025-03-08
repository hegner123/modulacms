CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,
    route_id      INTEGER NOT NULL
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    datatype_id   INTEGER NOT NULL
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

