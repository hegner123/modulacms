CREATE TABLE field (
    field_id INTEGER PRIMARY KEY,
    routeid INTEGER NOT NULL DEFAULT 1,
    parentid INTEGER, 
    label TEXT NOT NULL DEFAULT "unlabeled",
    data TEXT NOT NULL DEFAULT "",
    type TEXT NOT NULL DEFAULT "text",
    author TEXT NOT NULL DEFAULT "system",
    authorid INTEGER NOT NULL DEFAULT 1,
    datecreated TEXT,
    datemodified TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (author_id) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (route_id) REFERENCES route (route_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES datatype (datatype_id) ON DELETE SET NULL ON UPDATE CASCADE
);
