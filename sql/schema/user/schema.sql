CREATE TABLE user(
    user_id INTEGER PRIMARY KEY,
    datecreated TEXT NOT NULL,
    datemodified TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
);
