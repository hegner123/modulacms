-- name: DropValidationTable :exec
DROP TABLE IF EXISTS validations;

-- name: CreateValidationTable :exec
CREATE TABLE IF NOT EXISTS validations (
    validation_id TEXT PRIMARY KEY NOT NULL CHECK (length(validation_id) = 26),
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    config        TEXT NOT NULL DEFAULT '{}',
    author_id     TEXT
        REFERENCES users
            ON DELETE SET NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
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
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

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
