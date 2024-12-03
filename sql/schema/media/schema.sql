CREATE TABLE media(
    id INTEGER PRIMARY KEY,
    name  TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT NOT NULL DEFAULT "system",
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT NOT NULL,
    date_modified TEXT NOT NULL,
    url TEXT UNIQUE,
    mimetype TEXT,
    dimensions TEXT,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultrawide TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (author_id) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

