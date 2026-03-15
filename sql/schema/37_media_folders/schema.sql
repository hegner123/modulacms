CREATE TABLE IF NOT EXISTS media_folders (
    folder_id     TEXT PRIMARY KEY NOT NULL CHECK (length(folder_id) = 26),
    name          TEXT NOT NULL,
    parent_id     TEXT NULL REFERENCES media_folders(folder_id) ON DELETE RESTRICT,
    date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_media_folders_parent ON media_folders(parent_id);
