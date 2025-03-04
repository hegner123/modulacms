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

