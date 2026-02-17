CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_dt_datatypes_parent
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_dt_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX idx_datatypes_author ON datatypes(author_id);
