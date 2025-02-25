CREATE TABLE IF NOT EXISTS admin_datatypes(
    admin_dt_id    INTEGER
        primary key,
    admin_route_id INTEGER default NULL
        references admin_routes
            on update cascade on delete set default,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    not null,
    type           TEXT    not null,
    author         TEXT    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER not null
        references users (user_id)
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
);
