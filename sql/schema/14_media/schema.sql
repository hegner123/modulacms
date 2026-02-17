CREATE TABLE IF NOT EXISTS media (
    media_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(media_id) = 26),
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
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_author ON media(author_id);

CREATE TRIGGER IF NOT EXISTS update_media_modified
    AFTER UPDATE ON media
    FOR EACH ROW
    BEGIN
        UPDATE media SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE media_id = NEW.media_id;
    END;
