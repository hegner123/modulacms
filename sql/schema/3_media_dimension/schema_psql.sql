CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id TEXT PRIMARY KEY NOT NULL,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
