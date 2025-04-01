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
INSERT INTO tables (label, author_id) VALUES ('admin_datatype_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_datatype', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_field', 1);
INSERT INTO tables (label, author_id) VALUES ('admin_route', 1);
INSERT INTO tables (label, author_id) VALUES ('content_data', 1);
INSERT INTO tables (label, author_id) VALUES ('content_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('datatype', 1);
INSERT INTO tables (label, author_id) VALUES ('datatype_fields', 1);
INSERT INTO tables (label, author_id) VALUES ('field', 1);
INSERT INTO tables (label, author_id) VALUES ('media', 1);
INSERT INTO tables (label, author_id) VALUES ('media_dimension', 1);
INSERT INTO tables (label, author_id) VALUES ('permissions', 1);
INSERT INTO tables (label, author_id) VALUES ('role', 1);
INSERT INTO tables (label, author_id) VALUES ('route', 1);
INSERT INTO tables (label, author_id) VALUES ('session', 1);
INSERT INTO tables (label, author_id) VALUES ('table', 1);
INSERT INTO tables (label, author_id) VALUES ('token', 1);
INSERT INTO tables (label, author_id) VALUES ('user_oauth', 1);
INSERT INTO tables (label, author_id) VALUES ('user', 1);
