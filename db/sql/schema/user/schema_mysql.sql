CREATE TABLE user(
    id SERIAL PRIMARY KEY,
    datecreated TEXT,
    datemodified TEXT,
    username TEXT,
    name TEXT,
    email TEXT UNIQUE,
    hash TEXT,
    role TEXT
);
