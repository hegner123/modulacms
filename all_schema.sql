CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
);
CREATE TABLE IF NOT EXISTS admin_datatypes(
    admin_datatype_id    INTEGER
        primary key,
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
CREATE TABLE IF NOT EXISTS admin_fields
(
    admin_field_id INTEGER
        primary key,
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
CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL unique,
    permissions TEXT NOT NULL unique
);
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL
    REFERENCES users(username)
    ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
    REFERENCES users(user_id)
    ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id         INTEGER primary key,
    label         TEXT unique,
    width         INTEGER,
    height        INTEGER,
    aspect_ratio  TEXT
);
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,
    route_id      INTEGER NOT NULL
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    datatype_id   INTEGER NOT NULL
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT  DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id TEXT NOT NULL,  -- Unique identifier provided by the OAuth provider
    access_token TEXT,                     -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                    -- Optional: if token refresh is required
    token_expires_at TEXT,                 -- Optional: expiry time for the access token
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id       INTEGER NOT NULL
    REFERENCES routes(route_id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id       INTEGER NOT NULL
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_value         TEXT NOT NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP,
    history             TEXT
);

CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id   INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP,
    history TEXT
);
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL,
        references role
            on update cascade on delete set NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS fields
(
    field_id      INTEGER
        primary key,
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
    history TEXT
);
CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id      INTEGER NOT NULL
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_datatype_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT  DEFAULT NULL
);
-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id       INTEGER NOT NULL
    REFERENCES admin_routes(admin_route_id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id       INTEGER NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value         TEXT NOT NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP,
    history             TEXT
);
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
    mimetype             TEXT,
    dimensions           TEXT,
    url                  TEXT
        unique,
    optimized_mobile     TEXT,
    optimized_tablet     TEXT,
    optimized_desktop    TEXT,
    optimized_ultra_wide TEXT,
    author               TEXT    default "system" not null
    references users (username)
    on update cascade on delete set default,
    author_id            INTEGER default 1        not null
    references users (user_id)
    on update cascade on delete set default,
    date_created         TEXT    default CURRENT_TIMESTAMP,
    date_modified        TEXT    default CURRENT_TIMESTAMP
);
CREATE TABLE sessions (
    session_id   INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    created_at   TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at   TEXT,
    last_access  TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address   TEXT,
    user_agent   TEXT,
    session_data TEXT 
);

