CREATE TABLE IF NOT EXISTS media (
    media_id SERIAL PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT NOT NULL DEFAULT 'system',
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT UNIQUE,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT
);

