-- ModulaCMS Schema (MySQL)
-- Order follows sql/schema/ directory numbering
-- Generated from individual schema files

-- ===== 0_backups =====

CREATE TABLE IF NOT EXISTS backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    started_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at    TIMESTAMP NULL,
    duration_ms     INTEGER,
    record_count    BIGINT,
    size_bytes      BIGINT,
    replication_lsn VARCHAR(64),
    hlc_timestamp   BIGINT,
    storage_path    TEXT NOT NULL,
    checksum        VARCHAR(64),
    triggered_by    VARCHAR(64),
    error_message   TEXT,
    metadata        JSON,
    CONSTRAINT chk_backup_type CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    CONSTRAINT chk_backup_status CHECK (status IN ('started', 'completed', 'failed', 'verified'))
);

CREATE INDEX idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL,
    verified_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified_by      VARCHAR(64),
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL,
    error_message    TEXT,
    duration_ms      INTEGER,
    CONSTRAINT chk_verification_status CHECK (status IN ('passed', 'failed')),
    FOREIGN KEY (backup_id) REFERENCES backups(backup_id)
);

CREATE INDEX idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    date_created       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    hlc_timestamp    BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL,
    backup_ids       JSON NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT,
    CONSTRAINT chk_set_status CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_backup_sets_time ON backup_sets(date_created DESC);
CREATE INDEX idx_backup_sets_hlc ON backup_sets(hlc_timestamp);

-- ===== 0_change_events =====

CREATE TABLE IF NOT EXISTS change_events (
    event_id CHAR(26) PRIMARY KEY,
    hlc_timestamp BIGINT NOT NULL,
    wall_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,
    operation VARCHAR(20) NOT NULL,
    action VARCHAR(20),
    user_id CHAR(26),
    old_values JSON,
    new_values JSON,
    metadata JSON,
    request_id VARCHAR(255),
    ip VARCHAR(45),
    synced_at TIMESTAMP NULL,
    consumed_at TIMESTAMP NULL,
    CONSTRAINT chk_operation CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE'))
);

CREATE INDEX idx_events_record ON change_events(table_name, record_id);
CREATE INDEX idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX idx_events_node ON change_events(node_id);
CREATE INDEX idx_events_user ON change_events(user_id);
CREATE INDEX idx_events_unsynced ON change_events((synced_at IS NULL));
CREATE INDEX idx_events_unconsumed ON change_events((consumed_at IS NULL));

-- ===== 1_permissions =====

CREATE TABLE IF NOT EXISTS permissions (
    permission_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT perm_label_unique UNIQUE (label)
);

-- ===== 10_admin_fields =====

CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'text' NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_fields_admin_datatypes
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX idx_admin_fields_author ON admin_fields(author_id);

-- ===== 11_tokens =====

CREATE TABLE IF NOT EXISTS tokens (
    id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    token_type VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    revoked TINYINT(1) DEFAULT 0 NOT NULL,
    CONSTRAINT token
        UNIQUE (token),
    CONSTRAINT fk_tokens_users
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_tokens_user ON tokens(user_id);

-- ===== 12_user_oauth =====

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT user_oauth_ibfk_1
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_user_oauth_user ON user_oauth(user_id);

-- ===== 13_tables =====

CREATE TABLE IF NOT EXISTS tables (
    id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    author_id VARCHAR(26),
    CONSTRAINT label
        UNIQUE (label),
    CONSTRAINT fk_tables_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- ===== 14_media =====

CREATE TABLE IF NOT EXISTS media (
    media_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NULL,
    display_name TEXT NULL,
    alt TEXT NULL,
    caption TEXT NULL,
    description TEXT NULL,
    class TEXT NULL,
    mimetype TEXT NULL,
    dimensions TEXT NULL,
    url VARCHAR(255) NULL,
    srcset TEXT NULL,
    focal_x FLOAT NULL,
    focal_y FLOAT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT url
        UNIQUE (url),
    CONSTRAINT fk_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_media_author ON media(author_id);

-- ===== 15_sessions =====

CREATE TABLE sessions (
    session_id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_access TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    session_data TEXT NULL,
    CONSTRAINT fk_sessions_user_id
        FOREIGN KEY (user_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user ON sessions(user_id);

-- ===== 16_content_data =====

CREATE TABLE IF NOT EXISTS content_data (
    content_data_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    first_child_id VARCHAR(26) NULL,
    next_sibling_id VARCHAR(26) NULL,
    prev_sibling_id VARCHAR(26) NULL,
    route_id VARCHAR(26) NULL,
    datatype_id VARCHAR(26) NULL,
    author_id VARCHAR(26) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_data_datatypes
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_content_data_parent ON content_data(parent_id);
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX idx_content_data_author ON content_data(author_id);

-- ===== 17_content_fields =====

CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    route_id VARCHAR(26) NULL,
    content_data_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    field_value TEXT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_field_content_data
        FOREIGN KEY (content_data_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_fields
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_field_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_field_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_content_fields_route ON content_fields(route_id);
CREATE INDEX idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX idx_content_fields_field ON content_fields(field_id);
CREATE INDEX idx_content_fields_author ON content_fields(author_id);

-- ===== 18_admin_content_data =====

CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    first_child_id VARCHAR(26) NULL,
    next_sibling_id VARCHAR(26) NULL,
    prev_sibling_id VARCHAR(26) NULL,
    admin_route_id VARCHAR(26) NOT NULL,
    admin_datatype_id VARCHAR(26) NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES admin_content_data (admin_content_data_id)
             ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_datatypes
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX idx_admin_content_data_author ON admin_content_data(author_id);

-- ===== 19_admin_content_fields =====

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    admin_route_id VARCHAR(26) NULL,
    admin_content_data_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    admin_field_value TEXT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_field_admin_content_data
        FOREIGN KEY (admin_content_data_id) REFERENCES admin_content_data (admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_admin_route_id
        FOREIGN KEY (admin_route_id) REFERENCES admin_routes (admin_route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_field_fields
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_author_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX idx_admin_content_fields_author ON admin_content_fields(author_id);

-- ===== 2_roles =====

CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT label
        UNIQUE (label)
);

-- ===== 20_datatypes_fields =====

CREATE TABLE IF NOT EXISTS datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    datatype_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    CONSTRAINT fk_df_datatype
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_field
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_datatypes_fields_datatype ON datatypes_fields(datatype_id);
CREATE INDEX idx_datatypes_fields_field ON datatypes_fields(field_id);

-- ===== 21_admin_datatypes_fields =====

CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    admin_datatype_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    CONSTRAINT fk_df_admin_datatype
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_admin_field
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_admin_datatypes_fields_datatype ON admin_datatypes_fields(admin_datatype_id);
CREATE INDEX idx_admin_datatypes_fields_field ON admin_datatypes_fields(admin_field_id);

-- ===== 23_user_ssh_keys =====

-- user_ssh_keys table for storing SSH public keys linked to user accounts
CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id VARCHAR(26) PRIMARY KEY NOT NULL,
    user_id VARCHAR(26) NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL, -- "ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", etc.
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255), -- User-friendly label: "laptop", "work desktop", etc.
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Index for fast lookup by fingerprint during SSH auth
CREATE INDEX idx_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);

-- Index for listing user's keys
CREATE INDEX idx_ssh_keys_user_id ON user_ssh_keys(user_id);

-- ===== 24_content_relations =====

CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id VARCHAR(26) NOT NULL,
    source_content_id VARCHAR(26) NOT NULL,
    target_content_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_relation_id),
    CONSTRAINT fk_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_content_relations_unique UNIQUE (source_content_id, field_id, target_content_id)
);

CREATE INDEX idx_content_relations_target ON content_relations(target_content_id, date_created);
CREATE INDEX idx_content_relations_field ON content_relations(field_id);

-- ===== 25_admin_content_relations =====

CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_content_relation_id),
    CONSTRAINT fk_admin_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_admin_content_relations_unique UNIQUE (source_content_id, admin_field_id, target_content_id)
);

CREATE INDEX idx_admin_content_relations_target ON admin_content_relations(target_content_id, date_created);
CREATE INDEX idx_admin_content_relations_field ON admin_content_relations(admin_field_id);

-- ===== 26_role_permissions =====

CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(26) NOT NULL,
    role_id VARCHAR(26) NOT NULL,
    permission_id VARCHAR(26) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_permission FOREIGN KEY (permission_id) REFERENCES permissions(permission_id) ON DELETE CASCADE,
    CONSTRAINT uq_role_permission UNIQUE (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- ===== 3_media_dimension =====

CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NULL,
    width INT NULL,
    height INT NULL,
    aspect_ratio TEXT NULL,
    CONSTRAINT label
        UNIQUE (label)
);

-- ===== 4_users =====

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(26) PRIMARY KEY NOT NULL,
    username VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hash TEXT NOT NULL,
    role VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT username
        UNIQUE (username),
    CONSTRAINT fk_users_role
        FOREIGN KEY (role) REFERENCES roles (role_id)
            ON UPDATE CASCADE ON DELETE RESTRICT
);

-- ===== 5_admin_routes =====

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id VARCHAR(26) PRIMARY KEY NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT slug
        UNIQUE (slug),
    CONSTRAINT fk_admin_routes_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_admin_routes_author ON admin_routes(author_id);

-- ===== 6_routes =====

CREATE TABLE IF NOT EXISTS routes (
    route_id VARCHAR(26) PRIMARY KEY NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT unique_slug
        UNIQUE (slug),
    CONSTRAINT fk_routes_routes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_routes_author ON routes(author_id);

-- ===== 7_datatypes =====

CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_dt_datatypes_parent
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_dt_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX idx_datatypes_author ON datatypes(author_id);

-- ===== 8_fields =====

CREATE TABLE IF NOT EXISTS fields (
    field_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label VARCHAR(255) DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    author_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_fields_datatypes
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_fields_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_fields_parent ON fields(parent_id);
CREATE INDEX idx_fields_author ON fields(author_id);

-- ===== 9_admin_datatypes =====

CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_admin_datatypes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_admin_datatypes_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX idx_admin_datatypes_author ON admin_datatypes(author_id);
