CREATE TABLE IF NOT EXISTS fields (
    field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    translatable TINYINT NOT NULL DEFAULT 0,
    roles TEXT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_fields_datatypes
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_fields_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_fields_parent ON fields(parent_id);
CREATE INDEX idx_fields_author ON fields(author_id);
