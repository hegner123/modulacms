CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id INTEGER
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
