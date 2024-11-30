CREATE TABLE media(
    id INTEGER PRIMARY KEY,
    name  TEXT,
    displayname TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT NOT NULL DEFAULT "system",
    authorid INTEGER NOT NULL DEFAULT 1,
    datecreated TEXT NOT NULL,
    datemodified TEXT NOT NULL,
    url TEXT UNIQUE,
    mimetype TEXT,
    dimensions TEXT,
    optimizedmobile TEXT,
    optimizedtablet TEXT,
    optimizeddesktop TEXT,
    optimizedultrawide TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

