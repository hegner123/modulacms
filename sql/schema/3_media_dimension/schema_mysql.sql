CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NULL,
    width INT NULL,
    height INT NULL,
    aspect_ratio TEXT NULL,
    CONSTRAINT label
        UNIQUE (label)
);
