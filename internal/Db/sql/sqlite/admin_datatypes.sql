CREATE TABLE admin_datatypes (
    admin_datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

CREATE INDEX admin_datatypes_parent_id_index
    ON admin_datatypes (parent_id);

