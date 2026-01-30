CREATE TABLE IF NOT EXISTS tables (
    id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    author_id VARCHAR(26),
    CONSTRAINT label
        UNIQUE (label),
    CONSTRAINT fk_tables_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);
