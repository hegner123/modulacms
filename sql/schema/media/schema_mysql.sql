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
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_media_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

