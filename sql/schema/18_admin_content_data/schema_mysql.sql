CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INT AUTO_INCREMENT
        PRIMARY KEY,
    parent_id INT NULL,
    first_child_id INT NULL,
    next_sibling_id INT NULL,
    prev_sibling_id INT NULL,
    admin_route_id INT NOT NULL,
    admin_datatype_id INT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_datatypes
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX idx_admin_content_data_author ON admin_content_data(author_id);
