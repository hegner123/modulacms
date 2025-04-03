CREATE TABLE IF NOT EXISTS tables (
    id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT
);

