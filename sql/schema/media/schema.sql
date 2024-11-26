CREATE TABLE media(
    id INTEGER PRIMARY KEY,
    name  TEXT,
    displayname TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT NOT NULL,
    authorid TEXT NOT NULL,
    datecreated TEXT NOT NULL,
    datemodified TEXT NOT NULL,
    url TEXT UNIQUE,
    mimetype TEXT,
    dimensions TEXT,
    optimizedmobile TEXT,
    optimizedtablet TEXT,
    optimizeddesktop TEXT,
    optimizedultrawide TEXT,
    FOREIGN KEY (author) REFERENCES user(name),
    FOREIGN KEY (authorid) REFERENCES user(id)
);

