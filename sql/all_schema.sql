-- ModulaCMS Schema (SQLite)
-- Order follows sql/create_order.md
-- Note: permissions table is not in create_order.md but is included first

CREATE TABLE IF NOT EXISTS permissions (
    permission_id INTEGER
        PRIMARY KEY,
    table_id INTEGER NOT NULL,
    mode INTEGER NOT NULL,
    label TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    permissions TEXT NOT NULL
        UNIQUE
);

CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id INTEGER
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER
        PRIMARY KEY,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 4
        REFERENCES roles
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);


CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
    REFERENCES users
    ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);

CREATE TABLE IF NOT EXISTS datatypes(
    datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);


CREATE TABLE admin_datatypes (
    admin_datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);


CREATE TABLE admin_fields (
    admin_field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);

CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT 0
);



CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    oauth_provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TEXT NOT NULL,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL,
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255),
    date_created TEXT NOT NULL,
    last_used TEXT,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_user_id ON user_ssh_keys(user_id);


CREATE TABLE IF NOT EXISTS tables (
    id INTEGER
        PRIMARY KEY,
    label TEXT NOT NULL
        UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT
);


CREATE TABLE IF NOT EXISTS media (
    media_id INTEGER
        PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT
        UNIQUE,
    srcset TEXT,
    author_id INTEGER DEFAULT 1 NOT NULL
    REFERENCES users
    ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    session_id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);



CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    first_child_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    next_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    prev_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);


CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER
        PRIMARY KEY,
    route_id INTEGER
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);

CREATE TABLE IF NOT EXISTS fields(
    field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

);

CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    first_child_id INTEGER,
    next_sibling_id INTEGER,
    prev_sibling_id INTEGER,
    admin_route_id INTEGER NOT NULL,
    admin_datatype_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes(admin_datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER
        PRIMARY KEY ,
    admin_route_id INTEGER,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL DEFAULT 0,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    FOREIGN KEY (admin_field_id) REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
        ON DELETE SET DEFAULT
);


CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON DELETE CASCADE
);

