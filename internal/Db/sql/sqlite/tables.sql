CREATE TABLE tables (
    id INTEGER
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT
);

INSERT INTO tables (id, label, author_id) VALUES (1, 'admin_content_data', 1);
INSERT INTO tables (id, label, author_id) VALUES (2, 'admin_content_fields', 1);
INSERT INTO tables (id, label, author_id) VALUES (3, 'admin_datatypes_fields', 1);
INSERT INTO tables (id, label, author_id) VALUES (4, 'admin_datatype', 1);
INSERT INTO tables (id, label, author_id) VALUES (5, 'admin_field', 1);
INSERT INTO tables (id, label, author_id) VALUES (6, 'admin_route', 1);
INSERT INTO tables (id, label, author_id) VALUES (7, 'content_data', 1);
INSERT INTO tables (id, label, author_id) VALUES (8, 'content_fields', 1);
INSERT INTO tables (id, label, author_id) VALUES (9, 'datatype', 1);
INSERT INTO tables (id, label, author_id) VALUES (10, 'datatypes_fields', 1);
INSERT INTO tables (id, label, author_id) VALUES (11, 'field', 1);
INSERT INTO tables (id, label, author_id) VALUES (12, 'media', 1);
INSERT INTO tables (id, label, author_id) VALUES (13, 'media_dimension', 1);
INSERT INTO tables (id, label, author_id) VALUES (14, 'permissions', 1);
INSERT INTO tables (id, label, author_id) VALUES (15, 'role', 1);
INSERT INTO tables (id, label, author_id) VALUES (16, 'route', 1);
INSERT INTO tables (id, label, author_id) VALUES (17, 'session', 1);
INSERT INTO tables (id, label, author_id) VALUES (18, 'table', 1);
INSERT INTO tables (id, label, author_id) VALUES (19, 'token', 1);
INSERT INTO tables (id, label, author_id) VALUES (20, 'user_oauth', 1);
INSERT INTO tables (id, label, author_id) VALUES (21, 'user', 1);
