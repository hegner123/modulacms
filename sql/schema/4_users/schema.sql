CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY NOT NULL CHECK (length(user_id) = 26),
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
        REFERENCES roles
            ON DELETE SET NULL,
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

