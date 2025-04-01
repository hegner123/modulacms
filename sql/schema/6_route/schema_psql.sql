CREATE TABLE IF NOT EXISTS routes (
    route_id SERIAL
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
