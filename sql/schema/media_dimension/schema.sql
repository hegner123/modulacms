CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id         INTEGER primary key,
    label         TEXT unique,
    width         INTEGER,
    height        INTEGER,
    aspect_ratio  TEXT
);
