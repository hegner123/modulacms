-- name: DropMediaDimensionTable :exec
DROP TABLE media_dimensions;

-- name: CreateMediaDimensionTable :exec
CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id VARCHAR(26) PRIMARY KEY NOT NULL,
    label VARCHAR(255) NULL,
    width INT NULL,
    height INT NULL,
    aspect_ratio TEXT NULL,
    CONSTRAINT label
        UNIQUE (label)
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

-- name: CreateMediaDimension :exec
INSERT INTO media_dimensions(
    md_id,
    label,
    width,
    height,
    aspect_ratio
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
);

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
