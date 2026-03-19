-- name: DropAdminValidationTable :exec
DROP TABLE IF EXISTS admin_validations;

-- name: CreateAdminValidationTable :exec
CREATE TABLE IF NOT EXISTS admin_validations (
    admin_validation_id TEXT PRIMARY KEY NOT NULL,
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    config              TEXT NOT NULL DEFAULT '{}',
    author_id           TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminValidation :one
SELECT COUNT(*)
FROM admin_validations;

-- name: GetAdminValidation :one
SELECT * FROM admin_validations
WHERE admin_validation_id = $1 LIMIT 1;

-- name: ListAdminValidation :many
SELECT * FROM admin_validations
ORDER BY name, admin_validation_id;

-- name: ListAdminValidationPaginated :many
SELECT * FROM admin_validations
ORDER BY name, admin_validation_id
LIMIT $1 OFFSET $2;

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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
) RETURNING *;

-- name: UpdateAdminValidation :exec
UPDATE admin_validations
SET name = $1,
    description = $2,
    config = $3,
    author_id = $4,
    date_created = $5,
    date_modified = $6
WHERE admin_validation_id = $7;

-- name: DeleteAdminValidation :exec
DELETE FROM admin_validations
WHERE admin_validation_id = $1;
