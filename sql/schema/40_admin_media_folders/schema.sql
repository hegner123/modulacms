CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_folder_id) = 26),
    name            TEXT NOT NULL,
    parent_id       TEXT NULL REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT,
    date_created    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_admin_media_folders_parent ON admin_media_folders(parent_id);
