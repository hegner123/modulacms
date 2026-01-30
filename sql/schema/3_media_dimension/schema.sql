CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id TEXT PRIMARY KEY NOT NULL CHECK (length(md_id) = 26),
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
