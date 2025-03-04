CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER
        PRIMARY KEY,
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
