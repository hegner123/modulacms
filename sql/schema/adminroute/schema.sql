CREATE TABLE adminroute (
    id INTEGER PRIMARY KEY,
    author TEXT, 
    authorid TEXT, 
    slug TEXT UNIQUE, 
    title TEXT, 
    status INTEGER, 
    datecreated INTEGER, 
    datemodified INTEGER, 
    content TEXT, 
    template TEXT
);

