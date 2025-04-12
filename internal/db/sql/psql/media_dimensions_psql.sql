CREATE TABLE media_dimensions (
    md_id SERIAL
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

ALTER TABLE media_dimensions
    OWNER TO modula_u;

