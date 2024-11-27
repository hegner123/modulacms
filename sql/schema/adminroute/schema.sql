CREATE TABLE adminroute (
    id INTEGER PRIMARY KEY,
    author TEXT NOT NULL DEFAULT "system", 
    authorid INTEGER NOT NULL DEFAULT "0", 
    slug TEXT UNIQUE NOT NULL, 
    title TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    datecreated TEXT NOT NULL, 
    datemodified TEXT NOT NULL, 
    template TEXT NOT NULL DEFAULT "page.html",
    FOREIGN KEY (author) REFERENCES user(username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user(id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

