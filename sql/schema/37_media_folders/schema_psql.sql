CREATE TABLE IF NOT EXISTS media_folders (
    folder_id     TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    parent_id     TEXT NULL REFERENCES media_folders(folder_id) ON DELETE RESTRICT,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_media_folders_parent ON media_folders(parent_id);
