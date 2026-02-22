CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    admin_route_id VARCHAR(26) NULL,
    admin_content_data_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_field_admin_content_data
        FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data (admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_field_fields
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX idx_admin_content_fields_author ON admin_content_fields(author_id);
