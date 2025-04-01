CREATE TABLE datatypes (
    datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

