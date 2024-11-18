CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY,
    datecreated TEXT,
    datemodified TEXT,
    username TEXT,
    name TEXT,
    email TEXT,
    hash TEXT,
    role TEXT
);
