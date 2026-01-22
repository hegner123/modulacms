CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER
        PRIMARY KEY,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 4
        REFERENCES roles
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS update_users_modified
    AFTER UPDATE ON users
    FOR EACH ROW
    BEGIN
        UPDATE users SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE user_id = NEW.user_id;
    END;

