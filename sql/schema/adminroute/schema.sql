CREATE TABLE adminroute (
    id INTEGER PRIMARY KEY,
    author TEXT, 
    authorid TEXT, 
    slug TEXT UNIQUE, 
    title TEXT, 
    status INTEGER, 
    datecreated TEXT, 
    datemodified TEXT, 
    content TEXT, 
    template TEXT,
    deleted INTEGER
);

