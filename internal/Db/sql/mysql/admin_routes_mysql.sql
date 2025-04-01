CREATE TABLE admin_routes (
    admin_route_id INT AUTO_INCREMENT
        PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL ON UPDATE CURRENT_TIMESTAMP(),
    history TEXT NULL,
    CONSTRAINT slug
        UNIQUE (slug),
    CONSTRAINT fk_admin_routes_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

