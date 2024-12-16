
-- name: GetMediaDimension :one
SELECT * FROM media_dimension
WHERE id = ? LIMIT 1;

-- name: ListMediaDimension :many
SELECT * FROM media_dimension 
ORDER BY label;

-- name: CreateMediaDimension :one
INSERT INTO media_dimension(
    label,
    width,
    height
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: UpdateMediaDimension :exec
UPDATE media_dimension
set label = ?,
    width = ?,
    height = ? 
WHERE id = ?;

-- name: DeleteMediaDimension :exec
DELETE FROM media_dimension
WHERE id = ?;
