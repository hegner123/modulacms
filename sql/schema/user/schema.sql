CREATE TABLE user(
    user_id INTEGER PRIMARY KEY,
    date_created TEXT NOT NULL,
    date_modified TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
);
