CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_admin_datatypes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_datatypes_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX idx_admin_datatypes_author ON admin_datatypes(author_id);
