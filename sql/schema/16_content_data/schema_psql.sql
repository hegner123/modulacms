CREATE TABLE IF NOT EXISTS content_data (
    content_data_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_parent_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id INTEGER
        CONSTRAINT fk_first_child_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id INTEGER
        CONSTRAINT fk_next_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id INTEGER
        CONSTRAINT fk_prev_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    route_id INTEGER
        CONSTRAINT fk_routes
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    datatype_id INTEGER
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
