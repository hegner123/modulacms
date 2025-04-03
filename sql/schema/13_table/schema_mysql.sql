CREATE TABLE IF NOT EXISTS tables (
    id INT AUTO_INCREMENT
        PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    CONSTRAINT label
        UNIQUE (label),
    CONSTRAINT fk_tables_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);
