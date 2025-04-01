CREATE TABLE media (
    media_id SERIAL
        PRIMARY KEY,
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
    author_id INTEGER DEFAULT 1 NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE media
    OWNER TO modula_u;

CREATE INDEX media_srcset_index
    ON media (srcset);

