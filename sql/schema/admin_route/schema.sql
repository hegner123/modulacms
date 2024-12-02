CREATE TABLE admin_route (
    admin_route_id INTEGER PRIMARY KEY,
    author TEXT NOT NULL DEFAULT "system", 
    author_id INTEGER NOT NULL DEFAULT "0", 
    slug TEXT UNIQUE NOT NULL, 
    title TEXT NOT NULL, 
    status INTEGER NOT NULL, 
    date_created TEXT,
    date_modified TEXT,
    template TEXT NOT NULL DEFAULT "modula_base.html",
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);

