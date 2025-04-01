CREATE TABLE IF NOT EXISTS permissions (
    permission_id INTEGER PRIMARY KEY,
    table_id INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id         INTEGER PRIMARY KEY,
    label         TEXT UNIQUE,
    width         INTEGER,
    height        INTEGER,
    aspect_ratio  TEXT
);

CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
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
);

CREATE TABLE IF NOT EXISTS media
(
    media_id             INTEGER
        PRIMARY KEY,
    name                 TEXT,
    display_name         TEXT,
    alt                  TEXT,
    caption              TEXT,
    description          TEXT,
    class                TEXT,
    author               TEXT    DEFAULT 'system' NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id            INTEGER DEFAULT 1        NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    date_created         TEXT    DEFAULT CURRENT_TIMESTAMP,
    date_modified        TEXT    DEFAULT CURRENT_TIMESTAMP,
    mimetype             TEXT,
    dimensions           TEXT,
    url                  TEXT
        UNIQUE,
    optimized_mobile     TEXT,
    optimized_tablet     TEXT,
    optimized_desktop    TEXT,
    optimized_ultra_wide TEXT
);

CREATE TABLE sessions (
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
);

CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT
);

CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    author TEXT DEFAULT 'system' NOT NULL
        REFERENCES users(username)
        ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users(user_id)
        ON DELETE SET DEFAULT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id TEXT NOT NULL,  -- UNIQUE identifier provided by the OAuth provider
    access_token TEXT,                     -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                    -- Optional: if token refresh is required
    token_expires_at TEXT,                 -- Optional: expiry time for the access token
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    
    author TEXT DEFAULT 'system' NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id   INTEGER
        PRIMARY KEY,
    route_id      INTEGER DEFAULT NULL
        REFERENCES routes
        ON DELETE SET DEFAULT,
    parent_id     INTEGER DEFAULT NULL
        REFERENCES datatypes
        ON DELETE SET DEFAULT,
    label         TEXT                     NOT NULL,
    type          TEXT                     NOT NULL,
    author        TEXT    DEFAULT 'system' NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id     INTEGER DEFAULT 1        NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    history TEXT,
    date_created  TEXT    DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS fields
(
    field_id      INTEGER
        PRIMARY KEY,
    route_id      INTEGER DEFAULT NULL
        REFERENCES routes
        ON DELETE SET DEFAULT,
    parent_id     INTEGER DEFAULT NULL
        REFERENCES datatypes
        ON DELETE SET DEFAULT,
    label         TEXT    DEFAULT 'unlabeled' NOT NULL,
    data          TEXT                        NOT NULL,
    type          TEXT                        NOT NULL,
    author        TEXT    DEFAULT 'system'    NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id     INTEGER DEFAULT 1           NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    history TEXT,
    date_created  TEXT    DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tokens (
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
);

CREATE TABLE IF NOT EXISTS admin_datatypes(
    admin_datatype_id    INTEGER
        PRIMARY KEY,
    admin_route_id INTEGER DEFAULT NULL
        REFERENCES admin_routes
        ON DELETE SET DEFAULT,
    parent_id      INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
        ON DELETE SET DEFAULT,
    label          TEXT    NOT NULL,
    type           TEXT    NOT NULL,
    author         TEXT    NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id      INTEGER NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    date_created   TEXT    DEFAULT CURRENT_TIMESTAMP,
    date_modified  TEXT    DEFAULT CURRENT_TIMESTAMP,
    history        TEXT
);

CREATE TABLE IF NOT EXISTS admin_fields
(
    admin_field_id INTEGER
        PRIMARY KEY,
    admin_route_id INTEGER DEFAULT 1
        REFERENCES admin_routes
        ON DELETE SET DEFAULT,
    parent_id      INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
        ON DELETE SET DEFAULT,
    label          TEXT    DEFAULT 'unlabeled' NOT NULL,
    data           TEXT    DEFAULT ''          NOT NULL,
    type           TEXT    DEFAULT 'text'      NOT NULL,
    author         TEXT    DEFAULT 'system'    NOT NULL
        REFERENCES users (username)
        ON DELETE SET DEFAULT,
    author_id      INTEGER DEFAULT 1           NOT NULL
        REFERENCES users (user_id)
        ON DELETE SET DEFAULT,
    date_created   TEXT    DEFAULT CURRENT_TIMESTAMP,
    date_modified  TEXT    DEFAULT CURRENT_TIMESTAMP,
    history        TEXT
);

CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id INTEGER PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    CONSTRAINT fk_df_admin_datatype FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_df_admin_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INTEGER PRIMARY KEY,
    datatype_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    CONSTRAINT fk_df_datatype FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_df_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS admin_content_data (
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
);

CREATE TABLE IF NOT EXISTS admin_content_fields (
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
);

CREATE TABLE IF NOT EXISTS content_data (
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
);

CREATE TABLE IF NOT EXISTS content_fields (
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
);

-- Create indexes for foreign keys
CREATE INDEX IF NOT EXISTS idx_content_fields_route_id
ON content_fields(route_id);

CREATE INDEX IF NOT EXISTS idx_content_fields_content_data_id
ON content_fields(content_data_id);

CREATE INDEX IF NOT EXISTS idx_content_fields_field_id
ON content_fields(field_id);
