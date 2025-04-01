CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id SERIAL
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
