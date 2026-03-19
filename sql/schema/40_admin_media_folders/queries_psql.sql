-- name: DropAdminMediaFolderTable :exec
DROP TABLE IF EXISTS admin_media_folders;

-- name: CreateAdminMediaFolderTable :exec
CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id TEXT PRIMARY KEY NOT NULL,
    name            TEXT NOT NULL,
    parent_id       TEXT NULL REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT,
    date_created    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified   TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- name: CountAdminMediaFolder :one
SELECT COUNT(*)
FROM admin_media_folders;

-- name: GetAdminMediaFolder :one
SELECT * FROM admin_media_folders
WHERE admin_folder_id = $1 LIMIT 1;

-- name: ListAdminMediaFolders :many
SELECT * FROM admin_media_folders
ORDER BY name ASC;

-- name: ListAdminMediaFoldersPaginated :many
SELECT * FROM admin_media_folders
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: ListAdminMediaFoldersByParent :many
SELECT * FROM admin_media_folders
WHERE parent_id = $1
ORDER BY name ASC;

-- name: ListAdminMediaFoldersAtRoot :many
SELECT * FROM admin_media_folders
WHERE parent_id IS NULL
ORDER BY name ASC;

-- name: GetAdminMediaFolderByNameAndParent :one
SELECT * FROM admin_media_folders
WHERE parent_id = $1 AND name = $2 LIMIT 1;

-- name: GetAdminMediaFolderByNameAtRoot :one
SELECT * FROM admin_media_folders
WHERE parent_id IS NULL AND name = $1 LIMIT 1;

-- name: CreateAdminMediaFolder :one
INSERT INTO admin_media_folders (
    admin_folder_id,
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

-- name: UpdateAdminMediaFolder :exec
UPDATE admin_media_folders
SET name = $1,
    parent_id = $2,
    date_modified = $3
WHERE admin_folder_id = $4;

-- name: DeleteAdminMediaFolder :exec
DELETE FROM admin_media_folders
WHERE admin_folder_id = $1;
