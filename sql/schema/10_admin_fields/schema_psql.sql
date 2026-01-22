CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT DEFAULT ''::TEXT NOT NULL,
    type TEXT DEFAULT 'text'::TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
