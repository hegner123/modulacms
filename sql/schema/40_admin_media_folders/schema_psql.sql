CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id TEXT PRIMARY KEY NOT NULL,
    name            TEXT NOT NULL,
    parent_id       TEXT NULL REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT,
    date_created    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified   TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_media_folders_parent ON admin_media_folders(parent_id);
