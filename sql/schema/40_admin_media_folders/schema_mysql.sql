CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name            VARCHAR(255) NOT NULL,
    parent_id       VARCHAR(26) NULL,
    date_created    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_admin_media_folders_parent FOREIGN KEY (parent_id) REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT
);
CREATE INDEX idx_admin_media_folders_parent ON admin_media_folders(parent_id);
