CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id VARCHAR(26) PRIMARY KEY NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL ON UPDATE CURRENT_TIMESTAMP(),

    CONSTRAINT slug
        UNIQUE (slug),
    CONSTRAINT fk_admin_routes_users_user_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_admin_routes_author ON admin_routes(author_id);
