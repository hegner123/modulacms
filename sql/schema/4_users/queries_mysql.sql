-- name: DropUserTable :exec
DROP TABLE users;

-- name: CreateUserTable :exec
CREATE TABLE IF NOT EXISTS users (
    user_id INT AUTO_INCREMENT
        PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    hash TEXT NOT NULL,
    role INT NOT NULL DEFAULT 4,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP(),
    CONSTRAINT username
        UNIQUE (username),
    CONSTRAINT fk_users_role
        FOREIGN KEY (role) REFERENCES roles (role_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);

-- name: CreateUsersEmailIndex :exec
CREATE INDEX idx_users_email ON users(email);

-- name: CountUser :one
SELECT COUNT(*)
FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = ? LIMIT 1;


-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: GetUserId :one
SELECT user_id FROM users
WHERE email = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM users 
ORDER BY user_id ;

-- name: CreateUser :exec
INSERT INTO users (
    username,
    name,
    email,
    hash,
    role,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: GetLastUser :one
 SELECT * FROM users WHERE user_id = LAST_INSERT_ID();

-- name: UpdateUser :exec
UPDATE users
    SET username = ?,
        name = ?,
        email = ?,
        hash = ?,
        role = ?,
        date_created = ?,
        date_modified = ?
WHERE user_id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;
