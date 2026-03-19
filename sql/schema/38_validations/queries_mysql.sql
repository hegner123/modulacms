-- name: DropValidationTable :exec
DROP TABLE IF EXISTS validations;

-- name: CreateValidationTable :exec
CREATE TABLE IF NOT EXISTS validations (
    validation_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name          VARCHAR(255) NOT NULL,
    description   TEXT NOT NULL,
    config        TEXT NOT NULL,
    author_id     VARCHAR(26) NULL,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_validations_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);

-- name: CountValidation :one
SELECT COUNT(*)
FROM validations;

-- name: GetValidation :one
SELECT * FROM validations
WHERE validation_id = ? LIMIT 1;

-- name: ListValidation :many
SELECT * FROM validations
ORDER BY name, validation_id;

-- name: ListValidationPaginated :many
SELECT * FROM validations
ORDER BY name, validation_id
LIMIT ? OFFSET ?;

-- name: ListValidationsByName :many
SELECT * FROM validations
WHERE name LIKE CONCAT('%', sqlc.arg(name), '%')
ORDER BY name, validation_id;

-- name: CreateValidation :exec
INSERT INTO validations (
    validation_id,
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

-- name: UpdateValidation :exec
UPDATE validations
SET name = ?,
    description = ?,
    config = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE validation_id = ?;

-- name: DeleteValidation :exec
DELETE FROM validations
WHERE validation_id = ?;
