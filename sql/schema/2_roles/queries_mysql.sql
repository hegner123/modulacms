-- name: DropRoleTable :exec
DROP TABLE roles;

-- name: CreateRoleTable :exec
CREATE TABLE IF NOT EXISTS roles (
    role_id INT AUTO_INCREMENT
        PRIMARY KEY,
    label VARCHAR(255) NOT NULL,
    permissions LONGTEXT COLLATE utf8mb4_bin NULL
        CHECK (JSON_VALID(`permissions`)),
    CONSTRAINT label
        UNIQUE (label)
);

-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = ? LIMIT 1;

-- name: CountRole :one
SELECT COUNT(*)
FROM roles;

-- name: ListRole :many
SELECT * FROM roles;

-- name: CreateRole :exec
INSERT INTO roles (label, permissions) VALUES (?,?);

-- name: GetLastRole :one
SELECT * FROM roles WHERE role_id = LAST_INSERT_ID();

-- name: UpdateRole :exec
UPDATE roles
set label=?,
    permissions=?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
