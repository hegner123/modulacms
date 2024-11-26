CREATE TABLE route (
    id INTEGER PRIMARY KEY,
    author TEXT NOT NULL, 
    authorid TEXT NOT NULL, 
    slug TEXT UNIQUE NOT NULL, 
    title TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    datecreated TEXT NOT NULL, 
    datemodified TEXT NOT NULL, 
    content TEXT,
    FOREIGN KEY (author) REFERENCES user(name),
    FOREIGN KEY (authorid) REFERENCES user(id)
);
