CREATE TABLE fields (
    id INTEGER PRIMARY KEY,
    routeid INTEGER,
    author TEXT,
    authorid TEXT,
    key TEXT,
    type TEXT,
    data TEXT,
    datecreated TEXT,
    datemodified TEXT,
    componentid INTEGER,
    tags TEXT,
    parent TEXT,
    FOREIGN KEY (componentid) REFERENCES elements(id)
);


