CREATE TABLE datatype (
    datatype_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL DEFAULT 1,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT "system",
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT, 
    date_modified TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (author_id) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (route_id) REFERENCES route( route_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES datatype( datatype_id) ON DELETE SET NULL ON UPDATE CASCADE
);
