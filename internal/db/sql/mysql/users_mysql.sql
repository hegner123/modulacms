CREATE TABLE users (
    user_id INT AUTO_INCREMENT
        PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hash TEXT NOT NULL,
    role INT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP() NOT NULL ON UPDATE CURRENT_TIMESTAMP(),
    CONSTRAINT username
        UNIQUE (username),
    CONSTRAINT fk_users_role
        FOREIGN KEY (role) REFERENCES roles (role_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

INSERT INTO modula_db.users (username, name, email, hash, role, date_created, date_modified) VALUES ('admin', 'admin', 'admin@modulacms.com', 'asldjkf', 1, '2025-03-30 14:07:09', '2025-03-30 14:07:09');
