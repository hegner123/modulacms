CREATE TABLE admin_datatype (
    admin_dt_id INTEGER PRIMARY KEY,
    adminrouteid INTEGER NOT NULL DEFAULT 1,
    parentid INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL ,
    authorid INTEGER NOT NULL ,
    datecreated TEXT  DEFAULT (DATETIME('now', 'localtime')) ,
    datemodified TEXT DEFAULT (DATETIME('now', 'localtime')),
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (adminrouteid) REFERENCES adminroute(admin_route_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (parentid) REFERENCES admin_datatype(admin_dt_id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);
