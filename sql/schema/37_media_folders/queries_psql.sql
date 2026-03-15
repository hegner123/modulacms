-- name: DropMediaFolderTable :exec
DROP TABLE IF EXISTS media_folders;

-- name: CreateMediaFolderTable :exec
CREATE TABLE IF NOT EXISTS media_folders (
    folder_id     TEXT PRIMARY KEY NOT NULL,
    name          TEXT NOT NULL,
    parent_id     TEXT NULL REFERENCES media_folders(folder_id) ON DELETE RESTRICT,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- name: CountMediaFolder :one
SELECT COUNT(*)
FROM media_folders;

-- name: GetMediaFolder :one
SELECT * FROM media_folders
WHERE folder_id = $1 LIMIT 1;

-- name: ListMediaFolders :many
SELECT * FROM media_folders
ORDER BY name ASC;

-- name: ListMediaFoldersPaginated :many
SELECT * FROM media_folders
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: ListMediaFoldersByParent :many
SELECT * FROM media_folders
WHERE parent_id = $1
ORDER BY name ASC;

-- name: ListMediaFoldersAtRoot :many
SELECT * FROM media_folders
WHERE parent_id IS NULL
ORDER BY name ASC;

-- name: GetMediaFolderByNameAndParent :one
SELECT * FROM media_folders
WHERE parent_id = $1 AND name = $2 LIMIT 1;

-- name: GetMediaFolderByNameAtRoot :one
SELECT * FROM media_folders
WHERE parent_id IS NULL AND name = $1 LIMIT 1;

-- name: CreateMediaFolder :one
INSERT INTO media_folders (
    folder_id,
    name,
    parent_id,
    date_created,
    date_modified
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: UpdateMediaFolder :exec
UPDATE media_folders
SET name = $1,
    parent_id = $2,
    date_modified = $3
WHERE folder_id = $4;

-- name: DeleteMediaFolder :exec
DELETE FROM media_folders
WHERE folder_id = $1;
