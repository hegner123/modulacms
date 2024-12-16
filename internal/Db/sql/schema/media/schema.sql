CREATE TABLE IF NOT EXISTS media
(
    media_id             INTEGER
        primary key,
    name                 TEXT,
    display_name         TEXT,
    alt                  TEXT,
    caption              TEXT,
    description          TEXT,
    class                TEXT,
    author               TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id            INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    date_created         TEXT    default CURRENT_TIMESTAMP,
    date_modified        TEXT    default CURRENT_TIMESTAMP,
    mimetype             TEXT,
    dimensions           TEXT,
    url                  TEXT
        unique,
    optimized_mobile     TEXT,
    optimized_tablet     TEXT,
    optimized_desktop    TEXT,
    optimized_ultra_wide TEXT
);
