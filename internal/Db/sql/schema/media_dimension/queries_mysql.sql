-- name: CreateMediaDimensionTable :exec
CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INT AUTO_INCREMENT PRIMARY KEY,
    label VARCHAR(255) UNIQUE,
    width INT,
    height INT,
    aspect_ratio TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetMediaDimension :one
SELECT * FROM media_dimensions
WHERE md_id = ? LIMIT 1;

-- name: CountMediaDimension :one
SELECT COUNT(*)
FROM media_dimensions;

-- name: ListMediaDimension :many
SELECT * FROM media_dimensions 
ORDER BY label;

-- name: CreateMediaDimension :exec
INSERT INTO media_dimensions(
    label,
    width,
    height,
    aspect_ratio
) VALUES (
  ?, ?, ?, ?
);

-- name: GetLastMediaDimension :one
SELECT * FROM media_dimensions WHERE md_id = LAST_INSERT_ID();

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
