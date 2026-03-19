CREATE TABLE IF NOT EXISTS admin_media (
    admin_media_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_media_id) = 26),
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
    REFERENCES users
    ON DELETE SET NULL,
    folder_id TEXT NULL
    REFERENCES admin_media_folders(admin_folder_id)
    ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_media_author ON admin_media(author_id);
CREATE INDEX IF NOT EXISTS idx_admin_media_folder ON admin_media(folder_id);

CREATE TRIGGER IF NOT EXISTS update_admin_media_modified
    AFTER UPDATE ON admin_media
    FOR EACH ROW
    BEGIN
        UPDATE admin_media SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_media_id = NEW.admin_media_id;
    END;
