CREATE TABLE field (
    id INTEGER PRIMARY KEY,
    routeid INTEGER NOT NULL,
    parentid INTEGER,
    label TEXT NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    struct TEXT,
    author TEXT,
    authorid TEXT,
    datecreated TEXT,
    datemodified TEXT,
    FOREIGN KEY (routeid) REFERENCES adminroutes(id),
    FOREIGN KEY (parentid) REFERENCES fields(id)
);
