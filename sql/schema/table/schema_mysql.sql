CREATE TABLE tables (
    id INT NOT NULL AUTO_INCREMENT,
    label VARCHAR(255) UNIQUE,
    author_id INT NOT NULL DEFAULT 1,
    PRIMARY KEY (id),
    CONSTRAINT fk_tables_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

