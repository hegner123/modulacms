CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id INT AUTO_INCREMENT
        PRIMARY KEY,
    parent_id INT NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    type VARCHAR(255) DEFAULT 'text' NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    history TEXT NULL,
    CONSTRAINT fk_admin_fields_admin_datatypes
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);
