CREATE TABLE IF NOT EXISTS admin_validations (
    admin_validation_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name                VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL,
    config              TEXT NOT NULL,
    author_id           VARCHAR(26) NULL,
    date_created        TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified       TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_admin_validations_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_validations_name ON admin_validations(name);
CREATE INDEX idx_admin_validations_author ON admin_validations(author_id);
