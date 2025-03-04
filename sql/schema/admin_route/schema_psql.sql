CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id SERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT NOT NULL DEFAULT 'system'
        REFERENCES users(username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER NOT NULL DEFAULT 1
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

