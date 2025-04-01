CREATE TABLE sqlite_master (
    type TEXT,
    name TEXT,
    tbl_name TEXT,
    rootpage INT,
    sql TEXT
);

INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'permissions', 'permissions', 2, 'CREATE TABLE permissions (
    permission_id INTEGER PRIMARY KEY,
    table_id INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'media_dimensions', 'media_dimensions', 3, 'CREATE TABLE media_dimensions
(
    md_id         INTEGER PRIMARY KEY,
    label         TEXT UNIQUE,
    width         INTEGER,
    height        INTEGER,
    aspect_ratio  TEXT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_media_dimensions_1', 'media_dimensions', 4, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'roles', 'roles', 5, 'CREATE TABLE roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL UNIQUE
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_roles_1', 'roles', 6, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_roles_2', 'roles', 7, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'users', 'users', 8, 'CREATE TABLE users (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL
        REFERENCES roles
        ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_users_1', 'users', 9, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'sessions', 'sessions', 12, 'CREATE TABLE sessions (
    session_id   INTEGER PRIMARY KEY,
    user_id      INTEGER NOT NULL
        REFERENCES users
        ON DELETE CASCADE,
    created_at   TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at   TEXT,
    last_access  TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address   TEXT,
    user_agent   TEXT,
    session_data TEXT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'tables', 'tables', 13, 'CREATE TABLE tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_tables_1', 'tables', 14, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'user_oauth', 'user_oauth', 17, 'CREATE TABLE user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,        -- e.g., ''google'', ''facebook''
    oauth_provider_user_id TEXT NOT NULL,  -- UNIQUE identifier provided by the OAuth provider
    access_token TEXT,                     -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                    -- Optional: if token refresh is required
    token_expires_at TEXT,                 -- Optional: expiry time for the access token
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON DELETE CASCADE
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'tokens', 'tokens', 24, 'CREATE TABLE tokens (
    id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users (user_id)
        ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_tokens_1', 'tokens', 25, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_datatypes_fields', 'admin_datatypes_fields', 30, 'CREATE TABLE admin_datatypes_fields (
    id INTEGER PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    CONSTRAINT fk_df_admin_datatype FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_df_admin_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'datatypes_fields', 'datatypes_fields', 26, 'CREATE TABLE datatypes_fields (
    id INTEGER PRIMARY KEY,
    datatype_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    CONSTRAINT fk_df_datatype FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_df_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON DELETE CASCADE
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_content_data', 'admin_content_data', 31, 'CREATE TABLE admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id      INTEGER NOT NULL
        REFERENCES admin_routes(admin_route_id)
        ON DELETE CASCADE,
    admin_datatype_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_content_fields', 'admin_content_fields', 32, 'CREATE TABLE admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id       INTEGER NOT NULL
    REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    admin_content_data_id       INTEGER NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    admin_field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE,
    admin_field_value         TEXT NOT NULL,
    history             TEXT,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'content_data', 'content_data', 33, 'CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    route_id      INTEGER NOT NULL
        REFERENCES routes(route_id)
        ON DELETE CASCADE,
    datatype_id   INTEGER NOT NULL
        REFERENCES datatypes(datatype_id)
        ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'content_fields', 'content_fields', 34, 'CREATE TABLE content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields(field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'idx_content_fields_route_id', 'content_fields', 35, 'CREATE INDEX idx_content_fields_route_id
ON content_fields(route_id)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'idx_content_fields_content_data_id', 'content_fields', 37, 'CREATE INDEX idx_content_fields_content_data_id
ON content_fields(content_data_id)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'idx_content_fields_field_id', 'content_fields', 38, 'CREATE INDEX idx_content_fields_field_id
ON content_fields(field_id)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_datatypes', 'admin_datatypes', 39, 'CREATE TABLE "admin_datatypes"
(
    admin_datatype_id INTEGER
        primary key,
    parent_id         INTEGER default NULL
        references admin_datatypes
            on delete set default,
    label             TEXT    not null,
    type              TEXT    not null,
    author_id         INTEGER not null
        references users
            on delete set default,
    date_created      TEXT    default CURRENT_TIMESTAMP,
    date_modified     TEXT    default CURRENT_TIMESTAMP,
    history           TEXT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'admin_datatypes_parent_id_index', 'admin_datatypes', 27, 'CREATE INDEX admin_datatypes_parent_id_index
    on admin_datatypes (parent_id)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_fields', 'admin_fields', 40, 'CREATE TABLE "admin_fields"
(
    admin_field_id INTEGER
        primary key,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on delete set default,
    label          TEXT    default ''unlabeled'' not null,
    data           TEXT    default ''''          not null,
    type           TEXT    default ''text''      not null,
    author_id      INTEGER default 1           not null
        references users
            on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'admin_fields_parent_id_index', 'admin_fields', 28, 'CREATE INDEX admin_fields_parent_id_index
    on admin_fields (parent_id)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'admin_routes', 'admin_routes', 41, 'CREATE TABLE "admin_routes"
(
    admin_route_id INTEGER
        primary key,
    slug           TEXT              not null
        unique,
    title          TEXT              not null,
    status         INTEGER           not null,
    author_id      INTEGER default 1 not null
        references users
            on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_admin_routes_1', 'admin_routes', 42, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'fields', 'fields', 22, 'CREATE TABLE "fields"
(
    field_id      INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on delete set default,
    label         TEXT    default ''unlabeled'' not null,
    data          TEXT                        not null,
    type          TEXT                        not null,
    author_id     INTEGER default 1           not null
        references users
            on delete set default,
    history       TEXT,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'datatypes', 'datatypes', 23, 'CREATE TABLE "datatypes"
(
    datatype_id   INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on delete set default,
    label         TEXT                     not null,
    type          TEXT                     not null,
    ""            TEXT    default ''system'' not null,
    author_id     INTEGER default 1        not null
        references users
            on delete set default,
    history       TEXT,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'media', 'media', 18, 'CREATE TABLE "media"
(
    media_id             INTEGER
        primary key,
    name                 TEXT,
    display_name         TEXT,
    alt                  TEXT,
    caption              TEXT,
    description          TEXT,
    class                TEXT,
    author_id            INTEGER default 1 not null
        references users
            on delete set default,
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
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_media_1', 'media', 29, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('table', 'routes', 'routes', 10, 'CREATE TABLE "routes"
(
    route_id      INTEGER
        primary key,
    author_id     INTEGER   default 1 not null
        references users
            on delete set default,
    slug          TEXT                not null
        unique,
    title         TEXT                not null,
    status        INTEGER             not null,
    history       TEXT,
    date_created  TIMESTAMP default CURRENT_TIMESTAMP,
    date_modified TIMESTAMP default CURRENT_TIMESTAMP
)');
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'sqlite_autoindex_routes_1', 'routes', 11, null);
INSERT INTO sqlite_master (type, name, tbl_name, rootpage, sql) VALUES ('index', 'users_email_uindex', 'users', 15, 'CREATE UNIQUE INDEX users_email_uindex
    on users (email)');
