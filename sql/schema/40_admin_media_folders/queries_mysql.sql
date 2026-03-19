-- name: DropAdminMediaFolderTable :exec
DROP TABLE IF EXISTS admin_media_folders;

-- name: CreateAdminMediaFolderTable :exec
CREATE TABLE IF NOT EXISTS admin_media_folders (
    admin_folder_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name            VARCHAR(255) NOT NULL,
    parent_id       VARCHAR(26) NULL,
    date_created    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified   TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_admin_media_folders_parent FOREIGN KEY (parent_id) REFERENCES admin_media_folders(admin_folder_id) ON DELETE RESTRICT
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

-- name: CreateAdminMediaFolder :exec
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
);

-- name: UpdateAdminMediaFolder :exec
UPDATE admin_media_folders
SET name = ?,
    parent_id = ?,
    date_modified = ?
WHERE admin_folder_id = ?;

-- name: DeleteAdminMediaFolder :exec
DELETE FROM admin_media_folders
WHERE admin_folder_id = ?;
