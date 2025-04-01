CREATE TABLE admin_fields (
    admin_field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
