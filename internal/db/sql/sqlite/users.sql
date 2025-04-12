CREATE TABLE users (
    user_id INTEGER
        PRIMARY KEY,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL
        REFERENCES roles
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_email_uindex
    ON users (email);

INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES (1, 'admin', 'admin', 'admin@modulacms.com', 'saf', 1, '2025-03-30 15:08:40', '2025-03-30 15:08:40');
