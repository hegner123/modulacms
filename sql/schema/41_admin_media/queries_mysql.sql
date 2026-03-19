-- name: DropAdminMediaTable :exec
DROP TABLE admin_media;

-- name: CreateAdminMediaTable :exec
CREATE TABLE IF NOT EXISTS admin_media (
    admin_media_id VARCHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NULL,
    display_name TEXT NULL,
    alt TEXT NULL,
    caption TEXT NULL,
    description TEXT NULL,
    class TEXT NULL,
    mimetype TEXT NULL,
    dimensions TEXT NULL,
    url VARCHAR(255) NULL,
    srcset TEXT NULL,
    focal_x FLOAT NULL,
    focal_y FLOAT NULL,
    author_id VARCHAR(26) NOT NULL,
    folder_id VARCHAR(26) NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT admin_media_url
        UNIQUE (url),
    CONSTRAINT fk_admin_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_admin_media_admin_media_folders_folder_id
        FOREIGN KEY (folder_id) REFERENCES admin_media_folders (admin_folder_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);


-- name: CountAdminMedia :one
SELECT COUNT(*)
FROM admin_media;

-- name: GetAdminMedia :one
SELECT * FROM admin_media
WHERE admin_media_id = ? LIMIT 1;

-- name: GetAdminMediaByName :one
SELECT * FROM admin_media
WHERE name = ? LIMIT 1;

-- name: GetAdminMediaByUrl :one
SELECT * FROM admin_media
WHERE url = ? LIMIT 1;

-- name: ListAdminMedia :many
SELECT * FROM admin_media
ORDER BY name;

-- name: CreateAdminMedia :exec
INSERT INTO admin_media (
    admin_media_id,
    name,
    display_name,
    alt,
    caption,
    description,
    class,
    url,
    mimetype,
    dimensions,
    srcset,
    focal_x,
    focal_y,
    author_id,
    folder_id,
    date_created,
    date_modified
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
);
-- name: UpdateAdminMedia :exec
UPDATE admin_media
SET name = ?,
    display_name = ?,
    alt = ?,
    caption = ?,
    description = ?,
    class = ?,
    url = ?,
    mimetype = ?,
    dimensions = ?,
    srcset = ?,
    focal_x = ?,
    focal_y = ?,
    author_id = ?,
    folder_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_media_id = ?;

-- name: DeleteAdminMedia :exec
DELETE FROM admin_media
WHERE admin_media_id = ?;

-- name: ListAdminMediaPaginated :many
SELECT * FROM admin_media
ORDER BY name
LIMIT ? OFFSET ?;

-- name: ListAdminMediaByFolder :many
SELECT * FROM admin_media WHERE folder_id = ? ORDER BY date_created DESC;

-- name: ListAdminMediaByFolderPaginated :many
SELECT * FROM admin_media WHERE folder_id = ? ORDER BY date_created DESC LIMIT ? OFFSET ?;

-- name: ListAdminMediaUnfiled :many
SELECT * FROM admin_media WHERE folder_id IS NULL ORDER BY date_created DESC;

-- name: ListAdminMediaUnfiledPaginated :many
SELECT * FROM admin_media WHERE folder_id IS NULL ORDER BY date_created DESC LIMIT ? OFFSET ?;

-- name: CountAdminMediaByFolder :one
SELECT COUNT(*) FROM admin_media WHERE folder_id = ?;

-- name: CountAdminMediaUnfiled :one
SELECT COUNT(*) FROM admin_media WHERE folder_id IS NULL;

-- name: MoveAdminMediaToFolder :exec
UPDATE admin_media SET folder_id = ?, date_modified = ? WHERE admin_media_id = ?;
