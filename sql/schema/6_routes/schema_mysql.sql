CREATE TABLE IF NOT EXISTS routes (
    route_id VARCHAR(26) PRIMARY KEY NOT NULL,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT unique_slug
        UNIQUE (slug),
    CONSTRAINT fk_routes_routes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

CREATE INDEX idx_routes_author ON routes(author_id);
