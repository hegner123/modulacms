CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id INT AUTO_INCREMENT
        PRIMARY KEY,
    parent_id INT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_admin_datatypes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_admin_datatypes_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE
);

