CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
