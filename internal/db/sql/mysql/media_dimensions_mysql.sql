CREATE TABLE media_dimensions (
    md_id INT AUTO_INCREMENT
        PRIMARY KEY,
    label VARCHAR(255) NULL,
    width INT NULL,
    height INT NULL,
    aspect_ratio TEXT NULL,
    CONSTRAINT label
        UNIQUE (label)
);

