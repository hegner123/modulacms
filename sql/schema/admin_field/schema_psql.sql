CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id SERIAL PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'text',
    author TEXT NOT NULL DEFAULT 'system'
        REFERENCES users(username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER NOT NULL DEFAULT 1
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

