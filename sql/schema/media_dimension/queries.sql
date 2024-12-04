
-- name: GetMediaDimension :one
SELECT * FROM media_dimensions
WHERE md_id = ? LIMIT 1;

-- name: CountMD :one
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
  ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateMediaDimension :exec
UPDATE media_dimensions
set label = ?,
    width = ?,
    height = ?,
    aspect_ratio = ?
WHERE md_id = ?;

-- name: DeleteMediaDimension :exec
DELETE FROM media_dimensions
WHERE md_id = ?;
