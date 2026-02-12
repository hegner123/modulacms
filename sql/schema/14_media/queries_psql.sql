-- name: DropMediaTable :exec
DROP TABLE media;

-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media (
    media_id TEXT PRIMARY KEY NOT NULL,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT
        UNIQUE,
    srcset TEXT,
    author_id TEXT NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountMedia :one
SELECT COUNT(*)
FROM media;

-- name: GetMedia :one
SELECT * FROM media
WHERE media_id = $1 LIMIT 1;

-- name: GetMediaByName :one
SELECT * FROM media
WHERE name = $1 LIMIT 1;

-- name: GetMediaByUrl :one
SELECT * FROM media
WHERE url = $1 LIMIT 1;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY name;

-- name: CreateMedia :one
INSERT INTO media (
    media_id,
    name,
    display_name,
    alt,
    caption,
    description,
    class,
    url,
    mimetype,
    dimensions,
    srcset,
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
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14
)
RETURNING *;

-- name: UpdateMedia :exec
UPDATE media
SET name = $1,
    display_name = $2,
    alt = $3,
    caption = $4,
    description = $5,
    class = $6,
    url = $7,
    mimetype = $8,
    dimensions = $9,
    srcset = $10,
    author_id = $11,
    date_created = $12,
    date_modified = $13
WHERE media_id = $14;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = $1;

-- name: ListMediaPaginated :many
SELECT * FROM media
ORDER BY name
LIMIT $1 OFFSET $2;
