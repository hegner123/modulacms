CREATE TABLE IF NOT EXISTS media_folders (
    folder_id     VARCHAR(26) PRIMARY KEY NOT NULL,
    name          VARCHAR(255) NOT NULL,
    parent_id     VARCHAR(26) NULL,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_media_folders_parent FOREIGN KEY (parent_id) REFERENCES media_folders(folder_id) ON DELETE RESTRICT
);
CREATE INDEX idx_media_folders_parent ON media_folders(parent_id);
