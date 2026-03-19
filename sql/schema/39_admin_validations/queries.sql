-- name: DropAdminValidationTable :exec
DROP TABLE IF EXISTS admin_validations;

-- name: CreateAdminValidationTable :exec
CREATE TABLE IF NOT EXISTS admin_validations (
    admin_validation_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_validation_id) = 26),
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    config              TEXT NOT NULL DEFAULT '{}',
    author_id           TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminValidation :one
SELECT COUNT(*)
FROM admin_validations;

-- name: GetAdminValidation :one
SELECT * FROM admin_validations
WHERE admin_validation_id = ? LIMIT 1;

-- name: ListAdminValidation :many
SELECT * FROM admin_validations
ORDER BY name, admin_validation_id;

-- name: ListAdminValidationPaginated :many
SELECT * FROM admin_validations
ORDER BY name, admin_validation_id
LIMIT ? OFFSET ?;

-- name: ListAdminValidationsByName :many
SELECT * FROM admin_validations
WHERE name LIKE '%' || sqlc.arg(name) || '%'
ORDER BY name, admin_validation_id;

-- name: CreateAdminValidation :one
INSERT INTO admin_validations (
    admin_validation_id,
    name,
    description,
    config,
    author_id,
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
) RETURNING *;

-- name: UpdateAdminValidation :exec
UPDATE admin_validations
SET name = ?,
    description = ?,
    config = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_validation_id = ?;

-- name: DeleteAdminValidation :exec
DELETE FROM admin_validations
WHERE admin_validation_id = ?;
