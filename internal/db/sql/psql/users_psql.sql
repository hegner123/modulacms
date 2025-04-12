CREATE TABLE users (
    user_id SERIAL
        PRIMARY KEY,
    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER
        CONSTRAINT fk_users_role
            REFERENCES roles
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users
    OWNER TO modula_u;

CREATE INDEX users_email_index
    ON users (email);

INSERT INTO public.users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES (1, 'admin', 'admin', 'admin@modulacms.com', 'lskdfhj', 1, '2025-03-30 14:36:10.607129', '2025-03-30 14:36:10.607129');
