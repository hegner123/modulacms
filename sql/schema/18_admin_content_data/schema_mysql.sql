CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INT AUTO_INCREMENT
        PRIMARY KEY,
    admin_route_id INT NULL,
    parent_id INT NULL,
    admin_datatype_id INT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    history TEXT NULL,
    CONSTRAINT fk_admin_content_data_admin_datatypes
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_content_data (admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

