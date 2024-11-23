
CREATE TABLE user(
    id INTEGER PRIMARY KEY,
    datecreated TEXT,
    datemodified TEXT,
    username TEXT,
    name TEXT,
    email TEXT UNIQUE,
    hash TEXT,
    role TEXT
);
