CREATE TABLE IF NOT EXISTS fields (
    field_id INTEGER PRIMARY KEY,
    datatype_id       INTEGER NOT NULL
        REFERENCES admin_data(datatype_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_value         TEXT NOT NULL,
    author        TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP
);

