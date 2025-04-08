CREATE TABLE tables (
    id INT AUTO_INCREMENT
        PRIMARY KEY,
    label VARCHAR(255) NULL,
    author_id INT DEFAULT 1 NOT NULL,
    CONSTRAINT label
        UNIQUE (label),
    CONSTRAINT fk_tables_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_content_data', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_content_fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_datatypes_fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_datatypes', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('admin_routes', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('content_data', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('content_fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('datatypes', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('datatypes_fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('fields', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('media', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('media_dimensions', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('permissions', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('roles', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('routes', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('sessions', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('tables', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('tokens', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('user_oauth', 1);
INSERT INTO modula_db.tables (label, author_id) VALUES ('users', 1);
