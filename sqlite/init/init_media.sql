CREATE TABLE IF NOT EXISTS media(
    name  TEXT,
    displayname TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT,
    authorid INTEGER,
    datecreated TEXT,
    datemodified TEXT,
    url TEXT UNIQUE,
    mimetype TEXT,
    dimensions TEXT,
    optimizedmobile TEXT,
    optimizedtablet TEXT,
    optimizeddesktop TEXT,
    optimizedultrawide TEXT
);

