CREATE TABLE IF NOT EXISTS media (
    media_id INT AUTO_INCREMENT
        PRIMARY KEY,
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
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT url
        UNIQUE (url),
    CONSTRAINT fk_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

