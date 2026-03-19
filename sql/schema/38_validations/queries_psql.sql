-- name: DropValidationTable :exec
DROP TABLE IF EXISTS validations;

-- name: CreateValidationTable :exec
CREATE TABLE IF NOT EXISTS validations (
    validation_id TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    config        TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountValidation :one
SELECT COUNT(*)
FROM validations;

-- name: GetValidation :one
SELECT * FROM validations
WHERE validation_id = $1 LIMIT 1;

-- name: ListValidation :many
SELECT * FROM validations
ORDER BY name, validation_id;

-- name: ListValidationPaginated :many
SELECT * FROM validations
ORDER BY name, validation_id
LIMIT $1 OFFSET $2;

-- name: ListValidationsByName :many
SELECT * FROM validations
WHERE name LIKE '%' || sqlc.arg(name) || '%'
ORDER BY name, validation_id;

-- name: CreateValidation :one
INSERT INTO validations (
    validation_id,
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

-- name: UpdateValidation :exec
UPDATE validations
SET name = $1,
    description = $2,
    config = $3,
    author_id = $4,
    date_created = $5,
    date_modified = $6
WHERE validation_id = $7;

-- name: DeleteValidation :exec
DELETE FROM validations
WHERE validation_id = $1;
