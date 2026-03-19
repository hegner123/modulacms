-- name: DropAdminValidationTable :exec
DROP TABLE IF EXISTS admin_validations;

-- name: CreateAdminValidationTable :exec
CREATE TABLE IF NOT EXISTS admin_validations (
    admin_validation_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name                VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL,
    config              TEXT NOT NULL,
    author_id           VARCHAR(26) NULL,
    date_created        TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified       TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_admin_validations_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
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
WHERE name LIKE CONCAT('%', sqlc.arg(name), '%')
ORDER BY name, admin_validation_id;

-- name: CreateAdminValidation :exec
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
);

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
