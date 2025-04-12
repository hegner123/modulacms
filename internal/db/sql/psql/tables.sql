CREATE TABLE tables (
    id SERIAL
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);

ALTER TABLE tables
    OWNER TO modula_u;

INSERT INTO tables (label, author_id) VALUES ('admin_content_data', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_content_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_datatypes_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_datatypes', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_routes', 1);
INSERT INTO tables (label, author_id) VALUES ('content_data', 1);
INSERT INTO tables (label, author_id) VALUES ('content_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('datatypes', 1);
INSERT INTO tables (label, author_id) VALUES ('datatypes_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('fields', 1);
INSERT INTO tables (label, author_id) VALUES ('media', 1);
INSERT INTO tables (label, author_id) VALUES ('media_dimension', 1);
INSERT INTO tables (label, author_id) VALUES ('permissions', 1);
INSERT INTO tables (label, author_id) VALUES ('roles', 1);
INSERT INTO tables (label, author_id) VALUES ('routes', 1);
INSERT INTO tables (label, author_id) VALUES ('sessions', 1);
INSERT INTO tables (label, author_id) VALUES ('tables', 1);
INSERT INTO tables (label, author_id) VALUES ('tokens', 1);
INSERT INTO tables (label, author_id) VALUES ('user_oauth', 1);
INSERT INTO tables (label, author_id) VALUES ('users', 1);
