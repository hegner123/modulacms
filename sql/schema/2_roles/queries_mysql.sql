-- name: DropRoleTable :exec
DROP TABLE roles;

-- name: CreateRoleTable :exec
CREATE TABLE IF NOT EXISTS roles (
    role_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NOT NULL,
    system_protected BOOLEAN NOT NULL DEFAULT FALSE,
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
INSERT INTO roles (role_id, label, system_protected) VALUES (?,?,?);

-- name: UpdateRole :exec
UPDATE roles
SET label=?,
    system_protected=?
WHERE role_id = ?;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = ?;
