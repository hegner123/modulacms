-- ModulaCMS Schema (PostgreSQL)
-- Order follows sql/schema/ directory numbering
-- Generated from individual schema files

-- ===== 0_backups =====

CREATE TABLE IF NOT EXISTS backups (
    backup_id       CHAR(26) PRIMARY KEY,
    node_id         CHAR(26) NOT NULL,
    backup_type     VARCHAR(20) NOT NULL CHECK (backup_type IN ('full', 'incremental', 'snapshot')),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'verified')),
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at    TIMESTAMP WITH TIME ZONE,
    duration_ms     INTEGER,
    record_count    BIGINT,
    size_bytes      BIGINT,
    replication_lsn VARCHAR(64),
    hlc_timestamp   BIGINT,
    storage_path    TEXT NOT NULL,
    checksum        VARCHAR(64),
    triggered_by    VARCHAR(64),
    error_message   TEXT,
    metadata        JSONB
);

CREATE INDEX IF NOT EXISTS idx_backups_node ON backups(node_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_backups_hlc ON backups(hlc_timestamp);

CREATE TABLE IF NOT EXISTS backup_verifications (
    verification_id  CHAR(26) PRIMARY KEY,
    backup_id        CHAR(26) NOT NULL REFERENCES backups(backup_id),
    verified_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_by      VARCHAR(64),
    restore_tested   BOOLEAN DEFAULT FALSE,
    checksum_valid   BOOLEAN DEFAULT FALSE,
    record_count_match BOOLEAN DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL CHECK (status IN ('passed', 'failed')),
    error_message    TEXT,
    duration_ms      INTEGER
);

CREATE INDEX IF NOT EXISTS idx_verifications_backup ON backup_verifications(backup_id, verified_at DESC);

CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id    CHAR(26) PRIMARY KEY,
    date_created       TIMESTAMP WITH TIME ZONE NOT NULL,
    hlc_timestamp    BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed')),
    backup_ids       JSONB NOT NULL,
    node_count       INTEGER NOT NULL,
    completed_count  INTEGER DEFAULT 0,
    error_message    TEXT
);

CREATE INDEX IF NOT EXISTS idx_backup_sets_time ON backup_sets(date_created DESC);
CREATE INDEX IF NOT EXISTS idx_backup_sets_hlc ON backup_sets(hlc_timestamp);

-- ===== 0_change_events =====

CREATE TABLE IF NOT EXISTS change_events (
    event_id CHAR(26) PRIMARY KEY,
    hlc_timestamp BIGINT NOT NULL,
    wall_timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    node_id CHAR(26) NOT NULL,
    table_name VARCHAR(64) NOT NULL,
    record_id CHAR(26) NOT NULL,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action VARCHAR(20),
    user_id CHAR(26),
    old_values JSONB,
    new_values JSONB,
    metadata JSONB,
    request_id TEXT,
    ip TEXT,
    synced_at TIMESTAMP WITH TIME ZONE,
    consumed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_events_record ON change_events(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_events_hlc ON change_events(hlc_timestamp);
CREATE INDEX IF NOT EXISTS idx_events_node ON change_events(node_id);
CREATE INDEX IF NOT EXISTS idx_events_user ON change_events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_unsynced ON change_events(synced_at) WHERE synced_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_events_unconsumed ON change_events(consumed_at) WHERE consumed_at IS NULL;

-- ===== 1_permissions =====

CREATE TABLE IF NOT EXISTS permissions (
    permission_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL UNIQUE,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE
);

-- ===== 10_admin_fields =====

CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT DEFAULT ''::TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT DEFAULT 'text'::TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    translatable BOOLEAN NOT NULL DEFAULT FALSE,
    roles TEXT DEFAULT NULL,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_fields_parent ON admin_fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_fields_author ON admin_fields(author_id);

CREATE OR REPLACE FUNCTION update_admin_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_fields_modified_trigger
    BEFORE UPDATE ON admin_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_fields_modified();

-- ===== 11_tokens =====

CREATE TABLE IF NOT EXISTS tokens (
    id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL
        CONSTRAINT fk_tokens_users
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);

-- ===== 12_user_oauth =====

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    oauth_provider VARCHAR(255) NOT NULL,
    oauth_provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_user ON user_oauth(user_id);

-- ===== 13_tables =====

CREATE TABLE IF NOT EXISTS tables (
    id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL
        UNIQUE,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL
);

-- ===== 14_media =====

CREATE TABLE IF NOT EXISTS media (
    media_id TEXT PRIMARY KEY NOT NULL,
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
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_author ON media(author_id);

CREATE OR REPLACE FUNCTION update_media_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_media_modified_trigger
    BEFORE UPDATE ON media
    FOR EACH ROW
    EXECUTE FUNCTION update_media_modified();

-- ===== 15_sessions =====

CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL
        CONSTRAINT fk_sessions_user_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    last_access TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);


-- ===== 16_content_data =====

CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    route_id TEXT
        CONSTRAINT fk_routes
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    datatype_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    published_by TEXT
        CONSTRAINT fk_published_by
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    publish_at TIMESTAMP,
    revision INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_content_data_parent ON content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_content_data_route ON content_data(route_id);
CREATE INDEX IF NOT EXISTS idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX IF NOT EXISTS idx_content_data_author ON content_data(author_id);

CREATE OR REPLACE FUNCTION update_content_data_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_content_data_modified_trigger
    BEFORE UPDATE ON content_data
    FOR EACH ROW
    EXECUTE FUNCTION update_content_data_modified();

-- ===== 17_content_fields =====

CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id TEXT PRIMARY KEY NOT NULL,
    route_id TEXT
        CONSTRAINT fk_route_id
            REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id TEXT NOT NULL
        CONSTRAINT fk_content_data
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_fields_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_fields_route ON content_fields(route_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_content ON content_fields(content_data_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_field ON content_fields(field_id);
CREATE INDEX IF NOT EXISTS idx_content_fields_author ON content_fields(author_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cf_unique_locale ON content_fields(content_data_id, field_id, locale);

CREATE OR REPLACE FUNCTION update_content_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_content_fields_modified_trigger
    BEFORE UPDATE ON content_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_content_fields_modified();

-- ===== 18_admin_content_data =====

CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_route_id TEXT NOT NULL
        CONSTRAINT fk_admin_routes
            REFERENCES admin_routes
            ON UPDATE CASCADE,
    admin_datatype_id TEXT NOT NULL
        CONSTRAINT fk_admin_datatypes
            REFERENCES admin_datatypes
            ON UPDATE CASCADE,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    published_by TEXT
        CONSTRAINT fk_admin_published_by
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    publish_at TIMESTAMP,
    revision INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_admin_content_data_parent ON admin_content_data(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_route ON admin_content_data(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_datatype ON admin_content_data(admin_datatype_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_data_author ON admin_content_data(author_id);

CREATE OR REPLACE FUNCTION update_admin_content_data_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_content_data_modified_trigger
    BEFORE UPDATE ON admin_content_data
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_content_data_modified();

-- ===== 19_admin_content_fields =====

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id TEXT PRIMARY KEY NOT NULL,
    admin_route_id TEXT
        CONSTRAINT fk_admin_route_id
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id TEXT NOT NULL
        CONSTRAINT fk_admin_content_data
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        CONSTRAINT fk_admin_fields
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_content_fields_route ON admin_content_fields(admin_route_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_content ON admin_content_fields(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_field ON admin_content_fields(admin_field_id);
CREATE INDEX IF NOT EXISTS idx_admin_content_fields_author ON admin_content_fields(author_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_acf_unique_locale ON admin_content_fields(admin_content_data_id, admin_field_id, locale);

CREATE OR REPLACE FUNCTION update_admin_content_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_content_fields_modified_trigger
    BEFORE UPDATE ON admin_content_fields
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_content_fields_modified();

-- ===== 2_roles =====

CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL
        UNIQUE,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE
);


-- ===== 23_user_ssh_keys =====

-- user_ssh_keys table for storing SSH public keys linked to user accounts
CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id TEXT PRIMARY KEY NOT NULL,
    user_id TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL, -- "ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", etc.
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255), -- User-friendly label: "laptop", "work desktop", etc.
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
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
            ON UPDATE CASCADE ON DELETE CASCADE,
    target_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields(field_id)
            ON UPDATE CASCADE ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
            ON UPDATE CASCADE ON DELETE CASCADE,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        REFERENCES admin_fields(admin_field_id)
            ON UPDATE CASCADE ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
    role_id TEXT NOT NULL REFERENCES roles(role_id) ON UPDATE CASCADE ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions(permission_id) ON UPDATE CASCADE ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

-- ===== 27_field_types =====

CREATE TABLE IF NOT EXISTS field_types (
    field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);

-- ===== 28_admin_field_types =====

CREATE TABLE IF NOT EXISTS admin_field_types (
    admin_field_type_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_field_type_id) = 26),
    type TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL
);

-- ===== 29_plugins =====

CREATE TABLE IF NOT EXISTS plugins (
    plugin_id TEXT PRIMARY KEY NOT NULL CHECK (length(plugin_id) = 26),
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'installed',
    capabilities JSONB NOT NULL DEFAULT '[]',
    approved_access JSONB NOT NULL DEFAULT '{}',
    manifest_hash TEXT NOT NULL DEFAULT '',
    date_installed TIMESTAMPTZ NOT NULL,
    date_modified TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);

-- ===== 3_media_dimension =====

CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id TEXT PRIMARY KEY NOT NULL,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

-- ===== 30_pipelines =====

CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id TEXT PRIMARY KEY NOT NULL CHECK (length(pipeline_id) = 26),
    plugin_id TEXT NOT NULL REFERENCES plugins(plugin_id) ON UPDATE CASCADE ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    plugin_name TEXT NOT NULL,
    handler TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB NOT NULL DEFAULT '{}',
    date_created TIMESTAMPTZ NOT NULL,
    date_modified TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pipeline_unique ON pipelines(table_name, operation, plugin_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_plugin ON pipelines(plugin_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_table ON pipelines(table_name);

-- ===== 31_content_versions =====

CREATE TABLE IF NOT EXISTS content_versions (
    content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_version_id) = 26),
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    snapshot TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    label TEXT NOT NULL DEFAULT '',
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_by TEXT
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cv_content ON content_versions(content_data_id);
CREATE INDEX IF NOT EXISTS idx_cv_content_locale ON content_versions(content_data_id, locale);
CREATE INDEX IF NOT EXISTS idx_cv_published ON content_versions(content_data_id, locale) WHERE published = TRUE;

-- ===== 32_admin_content_versions =====

CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_version_id) = 26),
    admin_content_data_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    snapshot TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    label TEXT NOT NULL DEFAULT '',
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_by TEXT
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_acv_content ON admin_content_versions(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_acv_content_locale ON admin_content_versions(admin_content_data_id, locale);
CREATE INDEX IF NOT EXISTS idx_acv_published ON admin_content_versions(admin_content_data_id, locale) WHERE published = TRUE;

-- ===== 33_locales =====

CREATE TABLE IF NOT EXISTS locales (
    locale_id     TEXT PRIMARY KEY NOT NULL,
    code          TEXT NOT NULL UNIQUE,
    label         TEXT NOT NULL,
    is_default    BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    fallback_code TEXT,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_locales_code ON locales(code);

-- ===== 34_webhooks =====

CREATE TABLE IF NOT EXISTS webhooks (
    webhook_id    TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    url           TEXT NOT NULL,
    secret        TEXT NOT NULL DEFAULT '',
    events        TEXT NOT NULL DEFAULT '[]',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    headers       TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT NOT NULL REFERENCES users(user_id),
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active);

-- ===== 35_webhook_deliveries =====

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    delivery_id      TEXT PRIMARY KEY NOT NULL,
    webhook_id       TEXT NOT NULL REFERENCES webhooks(webhook_id) ON DELETE CASCADE,
    event            TEXT NOT NULL,
    payload          TEXT NOT NULL DEFAULT '{}',
    status           TEXT NOT NULL DEFAULT 'pending',
    attempts         INTEGER NOT NULL DEFAULT 0,
    last_status_code INTEGER,
    last_error       TEXT NOT NULL DEFAULT '',
    next_retry_at    TIMESTAMP,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    completed_at     TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_wd_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_wd_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_wd_retry ON webhook_deliveries(next_retry_at) WHERE status = 'retrying';

-- ===== 4_users =====

CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL
        CONSTRAINT fk_users_role
            REFERENCES roles
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_users_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_modified_trigger
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_users_modified();

-- ===== 5_admin_routes =====

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id TEXT PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_routes_author ON admin_routes(author_id);

CREATE OR REPLACE FUNCTION update_admin_routes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_routes_modified_trigger
    BEFORE UPDATE ON admin_routes
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_routes_modified();

-- ===== 6_routes =====

CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY NOT NULL,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_routes_author ON routes(author_id);

CREATE OR REPLACE FUNCTION update_routes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_routes_modified_trigger
    BEFORE UPDATE ON routes
    FOR EACH ROW
    EXECUTE FUNCTION update_routes_modified();

-- ===== 7_datatypes =====

CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE RESTRICT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_datatypes_parent ON datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_datatypes_author ON datatypes(author_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_datatypes_name ON datatypes(name);

CREATE OR REPLACE FUNCTION update_datatypes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_datatypes_modified_trigger
    BEFORE UPDATE ON datatypes
    FOR EACH ROW
    EXECUTE FUNCTION update_datatypes_modified();

-- ===== 8_fields =====

CREATE TABLE IF NOT EXISTS fields (
    field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'textarea', 'number', 'date', 'datetime', 'boolean', 'select', 'media', 'relation', 'json', 'richtext', 'slug', 'email', 'url')),
    translatable BOOLEAN NOT NULL DEFAULT FALSE,
    roles TEXT DEFAULT NULL,
    author_id TEXT
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fields_parent ON fields(parent_id);
CREATE INDEX IF NOT EXISTS idx_fields_author ON fields(author_id);

CREATE OR REPLACE FUNCTION update_fields_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_fields_modified_trigger
    BEFORE UPDATE ON fields
    FOR EACH ROW
    EXECUTE FUNCTION update_fields_modified();

-- ===== 9_admin_datatypes =====

CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE RESTRICT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_datatypes_parent ON admin_datatypes(parent_id);
CREATE INDEX IF NOT EXISTS idx_admin_datatypes_author ON admin_datatypes(author_id);

CREATE OR REPLACE FUNCTION update_admin_datatypes_modified()
RETURNS TRIGGER AS $$
BEGIN
    NEW.date_modified = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_admin_datatypes_modified_trigger
    BEFORE UPDATE ON admin_datatypes
    FOR EACH ROW
    EXECUTE FUNCTION update_admin_datatypes_modified();
