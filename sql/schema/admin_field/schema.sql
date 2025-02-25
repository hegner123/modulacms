CREATE TABLE IF NOT EXISTS admin_fields
(
    admin_field_id INTEGER
        primary key,
    admin_route_id INTEGER default 1
        references admin_routes
            on update cascade on delete set default,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    default "unlabeled" not null,
    data           TEXT    default ""          not null,
    type           TEXT    default "text"      not null,
    author         TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
);
