-- name: DropMediaDimensionTable :exec
DROP TABLE media_dimensions;

-- name: CreateMediaDimensionTable :exec
CREATE TABLE IF NOT EXISTS media_dimensions
(
    md_id INTEGER
        PRIMARY KEY,
    label TEXT
        UNIQUE,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

-- name: GetMediaDimension :one
SELECT * FROM media_dimensions
WHERE md_id = ? LIMIT 1;

-- name: CountMediaDimension :one
SELECT COUNT(*)
FROM media_dimensions;

-- name: ListMediaDimension :many
SELECT * FROM media_dimensions 
ORDER BY label;

-- name: CreateMediaDimension :one
INSERT INTO media_dimensions(
    label,    
    width, 
    height, 
    aspect_ratio
) VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateMediaDimension :exec
UPDATE media_dimensions
SET label = ?, 
    width = ?, 
    height = ?, 
    aspect_ratio = ?
WHERE md_id = ?;

-- name: DeleteMediaDimension :exec
DELETE FROM media_dimensions
WHERE md_id = ?;
