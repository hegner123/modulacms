CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation_id VARCHAR(26) NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'text' NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    translatable TINYINT NOT NULL DEFAULT 0,
    roles TEXT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_fields_admin_datatypes
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_admin_validations
        FOREIGN KEY (validation_id) REFERENCES admin_validations (admin_validation_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX idx_admin_fields_author ON admin_fields(author_id);
CREATE INDEX idx_admin_fields_validation ON admin_fields(validation_id);
