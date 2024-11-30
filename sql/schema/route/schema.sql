CREATE TABLE route (
    route_id INTEGER PRIMARY KEY,
    author TEXT NOT NULL DEFAULT "system", 
    authorid INTEGER NOT NULL DEFAULT 1, 
    slug TEXT UNIQUE NOT NULL, 
    title TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    datecreated TEXT NOT NULL, 
    datemodified TEXT NOT NULL, 
    content TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE 
);
