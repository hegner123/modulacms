CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    route_id VARCHAR(26) NULL,
    content_data_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    field_value TEXT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_field_content_data
        FOREIGN KEY (content_data_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_fields
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_field_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_content_fields_route ON content_fields(route_id);
CREATE INDEX idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX idx_content_fields_field ON content_fields(field_id);
CREATE INDEX idx_content_fields_author ON content_fields(author_id);
