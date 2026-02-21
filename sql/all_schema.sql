-- ModulaCMS Schema (SQLite)
-- Order follows sql/schema/ directory numbering
-- Generated from individual schema files

-- ===== 0_backups =====

CREATE TABLE IF NOT EXISTS backups (
    backup_id       TEXT PRIMARY KEY CHECK (length(backup_id) = 26),
    node_id         TEXT NOT NULL CHECK (length(node_id) = 26),
    backup_type     TEXT NOT NULL CHECK (backup_type IN ('full', 'incremental', 'differential')),
    status          TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    started_at      TEXT NOT NULL,
    completed_at    TEXT,
    duration_ms     INTEGER,
    record_count    INTEGER,
    size_bytes      INTEGER,
    replication_lsn TEXT,
    hlc_timestamp   INTEGER,
    storage_path    TEXT NOT NULL,
    checksum        TEXT,
    triggered_by    TEXT,
    error_message   TEXT,
    metadata        TEXT
);

CREATE INDEX IF NOT EXISTS idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  TEXT PRIMARY KEY CHECK (length(verification_id) = 26),
    backup_id        TEXT NOT NULL REFERENCES backups(backup_id),
    verified_at      TEXT NOT NULL,
    verified_by      TEXT,
    restore_tested   INTEGER DEFAULT 0,
    checksum_valid   INTEGER DEFAULT 0,
    record_count_match INTEGER DEFAULT 0,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'verified', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX IF NOT EXISTS idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    TEXT PRIMARY KEY CHECK (length(backup_set_id) = 26),
    date_created       TEXT NOT NULL,
    hlc_timestamp    INTEGER NOT NULL,
    status           TEXT NOT NULL CHECK (status IN ('pending', 'complete', 'partial')),
    backup_ids       TEXT NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX IF NOT EXISTS idx_backup_sets_time ON backup_sets(date_created DESC);
CREATE INDEX IF NOT EXISTS idx_backup_sets_hlc ON backup_sets(hlc_timestamp);

-- ===== 0_change_events =====

CREATE TABLE IF NOT EXISTS change_events (
    event_id TEXT PRIMARY KEY CHECK (length(event_id) = 26),
    hlc_timestamp INTEGER NOT NULL,
    wall_timestamp TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    node_id TEXT NOT NULL CHECK (length(node_id) = 26),
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL CHECK (length(record_id) = 26),
    operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action TEXT,
    user_id TEXT CHECK (user_id IS NULL OR length(user_id) = 26),
    old_values TEXT,
    new_values TEXT,
    metadata TEXT,
    request_id TEXT,
    ip TEXT,
    synced_at TEXT,
    consumed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_events_record ON change_events(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX IF NOT EXISTS idx_events_node ON change_events(node_id);
CREATE INDEX IF NOT EXISTS idx_events_user ON change_events(user_id);

-- ===== 1_permissions =====

CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL CHECK (length(permission_id) = 26),
    label TEXT NOT NULL UNIQUE,
    system_protected INTEGER NOT NULL DEFAULT 0
);

-- ===== 10_admin_fields =====

CREATE TABLE admin_fields (
    admin_field_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_fields_author ON admin_fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_fields_modified
    AFTER UPDATE ON admin_fields
    FOR EACH ROW
    BEGIN
        UPDATE admin_fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_field_id = NEW.admin_field_id;
    END;

-- ===== 11_tokens =====

CREATE TABLE IF NOT EXISTS tokens (
    id TEXT
        PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);


-- ===== 12_user_oauth =====

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(user_oauth_id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    oauth_provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TEXT NOT NULL,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_user ON user_oauth(user_id);

-- ===== 13_tables =====

CREATE TABLE IF NOT EXISTS tables (
    id TEXT
        PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL
);


-- ===== 14_media =====

CREATE TABLE IF NOT EXISTS media (
    media_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(media_id) = 26),
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
    focal_x REAL,
    focal_y REAL,
    author_id TEXT
    REFERENCES users
    ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_author ON media(author_id);

CREATE TRIGGER IF NOT EXISTS update_media_modified
    AFTER UPDATE ON media
    FOR EACH ROW
    BEGIN
        UPDATE media SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE media_id = NEW.media_id;
    END;

-- ===== 15_sessions =====

CREATE TABLE sessions (
    session_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(session_id) = 26),
    user_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);


-- ===== 16_content_data =====

CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_data_id) = 26),
    parent_id TEXT,
    first_child_id TEXT,
    next_sibling_id TEXT,
    prev_sibling_id TEXT,
    route_id TEXT NOT NULL,
    datatype_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_content_data_parent ON content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_content_data_route ON content_data(route_id);
CREATE INDEX IF NOT EXISTS idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX IF NOT EXISTS idx_content_data_author ON content_data(author_id);

CREATE TRIGGER IF NOT EXISTS update_content_data_modified
    AFTER UPDATE ON content_data
    FOR EACH ROW
    BEGIN
        UPDATE content_data SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE content_data_id = NEW.content_data_id;
    END;


-- ===== 17_content_fields =====

CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_field_id) = 26),
    route_id TEXT
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id TEXT NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_fields_route ON content_fields(route_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_field ON content_fields(field_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_author ON content_fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_content_fields_modified
    AFTER UPDATE ON content_fields
    FOR EACH ROW
    BEGIN
        UPDATE content_fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE content_field_id = NEW.content_field_id;
    END;

-- ===== 18_admin_content_data =====

CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_data_id) = 26),
    parent_id TEXT,
    first_child_id TEXT,
    next_sibling_id TEXT,
    prev_sibling_id TEXT,
    admin_route_id TEXT NOT NULL,
    admin_datatype_id TEXT NOT NULL,
    author_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data(admin_content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id) ON DELETE RESTRICT,
    FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes(admin_datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_author ON admin_content_data(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_content_data_modified
    AFTER UPDATE ON admin_content_data
    FOR EACH ROW
    BEGIN
        UPDATE admin_content_data SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_content_data_id = NEW.admin_content_data_id;
    END;

-- ===== 19_admin_content_fields =====

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_field_id) = 26),
    admin_route_id TEXT,
    admin_content_data_id TEXT NOT NULL,
    admin_field_id TEXT NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (admin_route_id) REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    FOREIGN KEY (admin_field_id) REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_author ON admin_content_fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_content_fields_modified
    AFTER UPDATE ON admin_content_fields
    FOR EACH ROW
    BEGIN
        UPDATE admin_content_fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_content_field_id = NEW.admin_content_field_id;
    END;

-- ===== 2_roles =====

CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL CHECK (length(role_id) = 26),
    label TEXT NOT NULL
        UNIQUE,
    system_protected INTEGER NOT NULL DEFAULT 0
);

-- ===== 20_datatypes_fields =====

CREATE TABLE IF NOT EXISTS datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_fields_field ON datatypes_fields(field_id);

-- ===== 21_admin_datatypes_fields =====

CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    admin_datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_fields_datatype ON admin_datatypes_fields(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_fields_field ON admin_datatypes_fields(admin_field_id);

-- ===== 23_user_ssh_keys =====

-- user_ssh_keys table for storing SSH public keys linked to user accounts
CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id TEXT PRIMARY KEY NOT NULL CHECK (length(ssh_key_id) = 26),
    user_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL, -- "ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", etc.
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255), -- User-friendly label: "laptop", "work desktop", etc.
    date_created TEXT NOT NULL,
    last_used TEXT,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Index for fast lookup by fingerprint during SSH auth
CREATE INDEX IF NOT EXISTS idx_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);

-- Index for listing user's keys
CREATE INDEX IF NOT EXISTS idx_ssh_keys_user_id ON user_ssh_keys(user_id);

-- ===== 24_content_relations =====

CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_relation_id) = 26),
    source_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    target_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields(field_id)
            ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- Unique constraint ordered to also serve ListBySourceAndField (prefix: source, field)
-- and ListBySource (prefix: source) queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_content_relations_unique
    ON content_relations(source_content_id, field_id, target_content_id);
-- Composite index for ListByTarget ORDER BY date_created
CREATE INDEX IF NOT EXISTS idx_content_relations_target
    ON content_relations(target_content_id, date_created);
-- Supports ON DELETE RESTRICT FK checks when deleting a field
CREATE INDEX IF NOT EXISTS idx_content_relations_field
    ON content_relations(field_id);

-- ===== 25_admin_content_relations =====

CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_relation_id) = 26),
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON DELETE CASCADE,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        REFERENCES admin_fields(admin_field_id)
            ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- Unique constraint ordered to also serve ListBySourceAndField (prefix: source, field)
-- and ListBySource (prefix: source) queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_content_relations_unique
    ON admin_content_relations(source_content_id, admin_field_id, target_content_id);
-- Composite index for ListByTarget ORDER BY date_created
CREATE INDEX IF NOT EXISTS idx_admin_content_relations_target
    ON admin_content_relations(target_content_id, date_created);
-- Supports ON DELETE RESTRICT FK checks when deleting a field
CREATE INDEX IF NOT EXISTS idx_admin_content_relations_field
    ON admin_content_relations(admin_field_id);

-- ===== 26_role_permissions =====

CREATE TABLE IF NOT EXISTS role_permissions (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    role_id TEXT NOT NULL REFERENCES roles ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

-- ===== 3_media_dimension =====

CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id TEXT PRIMARY KEY NOT NULL CHECK (length(md_id) = 26),
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

-- ===== 4_users =====

CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY NOT NULL CHECK (length(user_id) = 26),
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
        REFERENCES roles
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS update_users_modified
    AFTER UPDATE ON users
    FOR EACH ROW
    BEGIN
        UPDATE users SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE user_id = NEW.user_id;
    END;


-- ===== 5_admin_routes =====

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_route_id) = 26),
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_routes_author ON admin_routes(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_routes_modified
    AFTER UPDATE ON admin_routes
    FOR EACH ROW
    BEGIN
        UPDATE admin_routes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_route_id = NEW.admin_route_id;
    END;

-- ===== 6_routes =====

CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY NOT NULL CHECK (length(route_id) = 26),
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT
    REFERENCES users
    ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_routes_author ON routes(author_id);

CREATE TRIGGER IF NOT EXISTS update_routes_modified
    AFTER UPDATE ON routes
    FOR EACH ROW
    BEGIN
        UPDATE routes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE route_id = NEW.route_id;
    END;

-- ===== 7_datatypes =====

CREATE TABLE IF NOT EXISTS datatypes(
    datatype_id TEXT PRIMARY KEY NOT NULL CHECK (length(datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE RESTRICT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_author ON datatypes(author_id);

CREATE TRIGGER IF NOT EXISTS update_datatypes_modified
    AFTER UPDATE ON datatypes
    FOR EACH ROW
    BEGIN
        UPDATE datatypes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE datatype_id = NEW.datatype_id;
    END;

-- ===== 8_fields =====

CREATE TABLE IF NOT EXISTS fields(
    field_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fields_parent ON fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_fields_author ON fields(author_id);

CREATE TRIGGER IF NOT EXISTS update_fields_modified
    AFTER UPDATE ON fields
    FOR EACH ROW
    BEGIN
        UPDATE fields SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE field_id = NEW.field_id;
    END;

-- ===== 9_admin_datatypes =====

CREATE TABLE admin_datatypes (
    admin_datatype_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE RESTRICT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_author ON admin_datatypes(author_id);

CREATE TRIGGER IF NOT EXISTS update_admin_datatypes_modified
    AFTER UPDATE ON admin_datatypes
    FOR EACH ROW
    BEGIN
        UPDATE admin_datatypes SET date_modified = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
        WHERE admin_datatype_id = NEW.admin_datatype_id;
    END;
