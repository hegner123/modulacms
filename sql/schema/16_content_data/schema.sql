CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    first_child_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    next_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    prev_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);

