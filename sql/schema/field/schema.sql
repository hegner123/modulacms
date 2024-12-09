CREATE TABLE IF NOT EXISTS fields
(
    field_id      INTEGER
        primary key,
    route_id      INTEGER default NULL
        references routes
            on update cascade on delete set default,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT    default "unlabeled" not null,
    data          TEXT                        not null,
    type          TEXT                        not null,
    author        TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP,
    template      TEXT
);
