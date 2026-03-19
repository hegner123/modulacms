CREATE TABLE IF NOT EXISTS admin_media (
    admin_media_id TEXT PRIMARY KEY NOT NULL,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT
        UNIQUE,
    srcset TEXT,
    focal_x REAL,
    focal_y REAL,
    author_id TEXT
        CONSTRAINT fk_admin_media_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    folder_id TEXT NULL
        CONSTRAINT fk_admin_media_admin_media_folders_folder_id
            REFERENCES admin_media_folders(admin_folder_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_media_author ON admin_media(author_id);
CREATE INDEX IF NOT EXISTS idx_admin_media_folder ON admin_media(folder_id);

CREATE OR REPLACE FUNCTION update_admin_media_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_media_modified_trigger
    BEFORE UPDATE ON admin_media
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_media_modified();
