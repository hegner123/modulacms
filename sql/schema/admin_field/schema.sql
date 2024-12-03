CREATE TABLE admin_field (
    admin_field_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER NOT NULL DEFAULT 1,
    parent_id INTEGER NOT NULL DEFAULT 1,
    label TEXT NOT NULL DEFAULT "unlabeled",
    data TEXT NOT NULL DEFAULT "",
    type TEXT NOT NULL DEFAULT "text",
    author TEXT NOT NULL DEFAULT "system",
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT,
    date_modified TEXT,
    FOREIGN KEY (author) REFERENCES user (username) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (authorid) REFERENCES user (user_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (admin_route_id) REFERENCES admin_route (admin_route_id) ON DELETE SET DEFAULT ON UPDATE CASCADE,
    FOREIGN KEY (parentid) REFERENCES admin_datatype (admin_dt_id) ON DELETE SET DEFAULT ON UPDATE CASCADE
);
