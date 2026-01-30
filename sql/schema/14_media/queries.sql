-- name: DropMediaTable :exec
DROP TABLE media;

-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media(
    media_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(media_id) = 26),
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
    REFERENCES users
    ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CountMedia :one
SELECT COUNT(*)
FROM media;

-- name: GetMedia :one
SELECT * FROM media
WHERE media_id = ? LIMIT 1;

-- name: GetMediaByName :one
SELECT * FROM media
WHERE name = ? LIMIT 1;

-- name: GetMediaByUrl :one
SELECT * FROM media
WHERE url = ? LIMIT 1;

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
    mimetype,
    dimensions,
    url,
    srcset,
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
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateMedia :exec
UPDATE media
SET name = ?,
    display_name = ?,
    alt = ?,
    caption = ?,
    description = ?,
    class = ?,
    mimetype = ?,
    dimensions = ?,
    url = ?,
    srcset = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE media_id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = ?;
