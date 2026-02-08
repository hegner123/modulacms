CREATE TABLE IF NOT EXISTS permissions (
    permission_id INT PRIMARY KEY,
    table_id INT NOT NULL,
    mode INT NOT NULL,
    label VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) UNIQUE,
    width INT,
    height INT,
    aspect_ratio TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS roles (
    role_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    permissions JSON 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS users (
    user_id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hash TEXT NOT NULL,
    role INT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_users_role FOREIGN KEY (role)
        REFERENCES roles(role_id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS media (
    media_id INT AUTO_INCREMENT PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url VARCHAR(255) UNIQUE,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_media_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON DELETE RESTRICT,
    CONSTRAINT fk_media_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE sessions (
    session_id   INTEGER NOT NULL AUTO_INCREMENT,
    user_id      INTEGER NOT NULL, 
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP,
    last_access  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address   VARCHAR(45),
    user_agent   TEXT,
    session_data TEXT,
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tables (
    id INT NOT NULL AUTO_INCREMENT,
    label VARCHAR(255) UNIQUE,
    author_id INT NOT NULL DEFAULT 1,
    PRIMARY KEY (id),
    CONSTRAINT fk_tables_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS routes (
    route_id INT NOT NULL AUTO_INCREMENT,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (route_id),
    UNIQUE KEY UNIQUE_slug (slug),
    CONSTRAINT fk_routes_routes_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_routes_routes_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_oauth (
    user_oauth_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    oauth_provider VARCHAR(255) NOT NULL,        -- e.g., 'google', 'facebook'
    oauth_provider_user_id VARCHAR(255) NOT NULL,  -- UNIQUE identifier provided by the OAuth provider
    access_token TEXT,                             -- Optional: for making API calls on behalf of the user
    refresh_token TEXT,                            -- Optional: if token refresh is required
    token_expires_at TIMESTAMP NULL,               -- Optional: expiry time for the access token
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_ssh_keys (
    ssh_key_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    public_key TEXT NOT NULL,
    key_type VARCHAR(50) NOT NULL,
    fingerprint VARCHAR(255) NOT NULL UNIQUE,
    label VARCHAR(255),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_ssh_keys_fingerprint ON user_ssh_keys(fingerprint);
CREATE INDEX idx_ssh_keys_user_id ON user_ssh_keys(user_id);

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INT AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_routes_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT or SET NULL instead.
        ON DELETE RESTRICT,
    CONSTRAINT fk_admin_routes_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT instead.
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id INT AUTO_INCREMENT PRIMARY KEY,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_dt_datatypes_parent FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_dt_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON DELETE RESTRICT,
    CONSTRAINT fk_dt_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fields (
    field_id INT AUTO_INCREMENT PRIMARY KEY,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_fields_datatypes FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_fields_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON DELETE RESTRICT,
    CONSTRAINT fk_fields_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token_type VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked TINYINT(1) DEFAULT 0,
    CONSTRAINT fk_tokens_users FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id INT NOT NULL AUTO_INCREMENT,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (admin_datatype_id),
    CONSTRAINT fk_admin_datatypes_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_admin_datatypes_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_admin_datatypes_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id INT AUTO_INCREMENT PRIMARY KEY,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL, -- MySQL does not allow a DEFAULT value for TEXT
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type VARCHAR(255) NOT NULL DEFAULT 'text',
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_fields_admin_datatypes FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_admin_fields_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON DELETE RESTRICT,
    CONSTRAINT fk_admin_fields_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id INT AUTO_INCREMENT PRIMARY KEY,
    admin_datatype_id INT NOT NULL,
    admin_field_id INT NOT NULL,
    CONSTRAINT fk_admin_datatypes_fields_admin_datatype FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_admin_datatypes_fields_admin_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datatype_id INT NOT NULL,
    field_id INT NOT NULL,
    CONSTRAINT fk_datatypes_fields_datatype FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_datatypes_fields_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id    INT DEFAULT NULL,
    admin_datatype_id INT DEFAULT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_data_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_data_admin_datatypes FOREIGN KEY (admin_datatype_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id INT,
    admin_content_data_id INT NOT NULL,
    admin_field_id INT NOT NULL,
    admin_field_value TEXT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_content_field_admin_content_data FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_field_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_admin_content_field_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id    INT DEFAULT NULL,
    datatype_id INT DEFAULT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_data_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_content_data_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id INT,
    content_data_id INT NOT NULL,
    field_id INT NOT NULL,
    field_value TEXT NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_field_content_data FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_content_field_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON DELETE SET NULL,
    CONSTRAINT fk_content_field_fields FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id VARCHAR(26) NOT NULL,
    source_content_id VARCHAR(26) NOT NULL,
    target_content_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_relation_id),
    CONSTRAINT chk_content_relations_no_self_ref CHECK (source_content_id != target_content_id),
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_content_relations_target ON content_relations(target_content_id, date_created);
CREATE INDEX idx_content_relations_field ON content_relations(field_id);

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
    CONSTRAINT chk_admin_content_relations_no_self_ref CHECK (source_content_id != target_content_id),
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_admin_content_relations_target ON admin_content_relations(target_content_id, date_created);
CREATE INDEX idx_admin_content_relations_field ON admin_content_relations(admin_field_id);
