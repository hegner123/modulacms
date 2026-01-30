CREATE TABLE IF NOT EXISTS tables (
    id TEXT
        PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL
);

