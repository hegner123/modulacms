-- name: DropMediaFolderTable :exec
DROP TABLE IF EXISTS media_folders;

-- name: CreateMediaFolderTable :exec
CREATE TABLE IF NOT EXISTS media_folders (
    folder_id     TEXT PRIMARY KEY NOT NULL CHECK (length(folder_id) = 26),
    name          TEXT NOT NULL,
    parent_id     TEXT NULL REFERENCES media_folders(folder_id) ON DELETE RESTRICT,
    date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CountMediaFolder :one
SELECT COUNT(*)
FROM media_folders;

-- name: GetMediaFolder :one
SELECT * FROM media_folders
WHERE folder_id = ? LIMIT 1;

-- name: ListMediaFolders :many
SELECT * FROM media_folders
ORDER BY name ASC;

-- name: ListMediaFoldersPaginated :many
SELECT * FROM media_folders
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListMediaFoldersByParent :many
SELECT * FROM media_folders
WHERE parent_id = ?
ORDER BY name ASC;

-- name: ListMediaFoldersAtRoot :many
SELECT * FROM media_folders
WHERE parent_id IS NULL
ORDER BY name ASC;

-- name: GetMediaFolderByNameAndParent :one
SELECT * FROM media_folders
WHERE parent_id = ? AND name = ? LIMIT 1;

-- name: GetMediaFolderByNameAtRoot :one
SELECT * FROM media_folders
WHERE parent_id IS NULL AND name = ? LIMIT 1;

-- name: CreateMediaFolder :one
INSERT INTO media_folders (
    folder_id,
    name,
    parent_id,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateMediaFolder :exec
UPDATE media_folders
SET name = ?,
    parent_id = ?,
    date_modified = ?
WHERE folder_id = ?;

-- name: DeleteMediaFolder :exec
DELETE FROM media_folders
WHERE folder_id = ?;
