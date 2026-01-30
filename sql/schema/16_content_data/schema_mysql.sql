CREATE TABLE IF NOT EXISTS content_data (
    content_data_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    first_child_id VARCHAR(26) NULL,
    next_sibling_id VARCHAR(26) NULL,
    prev_sibling_id VARCHAR(26) NULL,
    route_id VARCHAR(26) NULL,
    datatype_id VARCHAR(26) NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_data_datatypes
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_content_data_parent ON content_data(parent_id);
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX idx_content_data_author ON content_data(author_id);
