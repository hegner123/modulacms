-- name: DropAdminMediaFolderTable :exec
DROP TABLE IF EXISTS admin_media_folders;

-- name: CreateAdminMediaFolderTable :exec
CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_folder_id) = 26),
    name            TEXT NOT NULL,
    parent_id       TEXT NULL REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT,
    date_created    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminMediaFolder :one
SELECT COUNT(*)
FROM admin_media_folders;

-- name: GetAdminMediaFolder :one
SELECT * FROM admin_media_folders
WHERE admin_folder_id = ? LIMIT 1;

-- name: ListAdminMediaFolders :many
SELECT * FROM admin_media_folders
ORDER BY name ASC;

-- name: ListAdminMediaFoldersPaginated :many
SELECT * FROM admin_media_folders
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: ListAdminMediaFoldersByParent :many
SELECT * FROM admin_media_folders
WHERE parent_id = ?
ORDER BY name ASC;

-- name: ListAdminMediaFoldersAtRoot :many
SELECT * FROM admin_media_folders
WHERE parent_id IS NULL
ORDER BY name ASC;

-- name: GetAdminMediaFolderByNameAndParent :one
SELECT * FROM admin_media_folders
WHERE parent_id = ? AND name = ? LIMIT 1;

-- name: GetAdminMediaFolderByNameAtRoot :one
SELECT * FROM admin_media_folders
WHERE parent_id IS NULL AND name = ? LIMIT 1;

-- name: CreateAdminMediaFolder :one
INSERT INTO admin_media_folders (
    admin_folder_id,
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

-- name: UpdateAdminMediaFolder :exec
UPDATE admin_media_folders
SET name = ?,
    parent_id = ?,
    date_modified = ?
WHERE admin_folder_id = ?;

-- name: DeleteAdminMediaFolder :exec
DELETE FROM admin_media_folders
WHERE admin_folder_id = ?;
