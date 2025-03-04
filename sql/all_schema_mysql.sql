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
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_dt_id INT NOT NULL AUTO_INCREMENT,
    admin_route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
    PRIMARY KEY (admin_dt_id),
    CONSTRAINT fk_admin_route_id FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_route_id INT NOT NULL DEFAULT 1,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL, -- MySQL does not allow a default value for TEXT
    type VARCHAR(255) NOT NULL DEFAULT 'text',
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_admin_routes FOREIGN KEY (admin_route_id)
        REFERENCES admin_routes(admin_route_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (parent_id)
        REFERENCES admin_datatypes(admin_datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS roles (
    role_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) NOT NULL UNIQUE,
    permissions JSON 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS routes (
    route_id INT NOT NULL AUTO_INCREMENT,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    history TEXT,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (route_id),
    UNIQUE KEY unique_slug (slug),
    CONSTRAINT fk_routes_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE
        ON DELETE NO ACTION,
    CONSTRAINT fk_routes_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) UNIQUE,
    width INT,
    height INT,
    aspect_ratio TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    admin_dt_id INT DEFAULT NULL,
    history TEXT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_datatypes FOREIGN KEY (admin_dt_id)
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INT AUTO_INCREMENT PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT,
    CONSTRAINT fk_users_username FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT or SET NULL instead.
        ON DELETE RESTRICT,
    CONSTRAINT fk_users_user_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE 
        -- ON DELETE SET DEFAULT is not supported in MySQL; consider using RESTRICT instead.
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INT AUTO_INCREMENT PRIMARY KEY,
    content_data_id INT NOT NULL,
    admin_field_id INT NOT NULL,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_content_data FOREIGN KEY (content_data_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_fields FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_datatypes_parent FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET NULL
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
        ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fields (
    field_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id INT DEFAULT NULL,
    parent_id INT DEFAULT NULL,
    label VARCHAR(255) NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_datatypes FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
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

CREATE TABLE IF NOT EXISTS media (
    media_id INT AUTO_INCREMENT PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author VARCHAR(255) NOT NULL DEFAULT 'system',
    author_id INT NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    mimetype TEXT,
    dimensions TEXT,
    url VARCHAR(255) UNIQUE,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

