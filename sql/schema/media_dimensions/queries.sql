
-- name: GetMediaDimension :one
SELECT * FROM media_dimensions
WHERE id = ? LIMIT 1;

-- name: ListMediaDimensions :many
SELECT * FROM media_dimensions 
ORDER BY label;

-- name: CreateMediaDimension :one
INSERT INTO media_dimensions(
    label,
    width,
    height
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: UpdateMediaDimension :exec
UPDATE media_dimensions
set label = ?,
    width = ?,
    height = ? 
WHERE id = ?;

-- name: DeleteMediaDimension :exec
DELETE FROM media_dimensions
WHERE id = ?;
