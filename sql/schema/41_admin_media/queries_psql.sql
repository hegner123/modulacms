-- name: DropAdminMediaTable :exec
DROP TABLE admin_media;

-- name: CreateAdminMediaTable :exec
CREATE TABLE IF NOT EXISTS admin_media (
    admin_media_id TEXT PRIMARY KEY NOT NULL,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT
        UNIQUE,
    srcset TEXT,
    focal_x REAL,
    focal_y REAL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_admin_media_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    folder_id TEXT NULL
        CONSTRAINT fk_admin_media_admin_media_folders_folder_id
            REFERENCES admin_media_folders(admin_folder_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminMedia :one
SELECT COUNT(*)
FROM admin_media;

-- name: GetAdminMedia :one
SELECT * FROM admin_media
WHERE admin_media_id = $1 LIMIT 1;

-- name: GetAdminMediaByName :one
SELECT * FROM admin_media
WHERE name = $1 LIMIT 1;

-- name: GetAdminMediaByUrl :one
SELECT * FROM admin_media
WHERE url = $1 LIMIT 1;

-- name: ListAdminMedia :many
SELECT * FROM admin_media
ORDER BY name;

-- name: CreateAdminMedia :one
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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17
)
RETURNING *;

-- name: UpdateAdminMedia :exec
UPDATE admin_media
SET name = $1,
    display_name = $2,
    alt = $3,
    caption = $4,
    description = $5,
    class = $6,
    url = $7,
    mimetype = $8,
    dimensions = $9,
    srcset = $10,
    focal_x = $11,
    focal_y = $12,
    author_id = $13,
    folder_id = $14,
    date_created = $15,
    date_modified = $16
WHERE admin_media_id = $17;

-- name: DeleteAdminMedia :exec
DELETE FROM admin_media
WHERE admin_media_id = $1;

-- name: ListAdminMediaPaginated :many
SELECT * FROM admin_media
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: ListAdminMediaByFolder :many
SELECT * FROM admin_media WHERE folder_id = $1 ORDER BY date_created DESC;

-- name: ListAdminMediaByFolderPaginated :many
SELECT * FROM admin_media WHERE folder_id = $1 ORDER BY date_created DESC LIMIT $2 OFFSET $3;

-- name: ListAdminMediaUnfiled :many
SELECT * FROM admin_media WHERE folder_id IS NULL ORDER BY date_created DESC;

-- name: ListAdminMediaUnfiledPaginated :many
SELECT * FROM admin_media WHERE folder_id IS NULL ORDER BY date_created DESC LIMIT $1 OFFSET $2;

-- name: CountAdminMediaByFolder :one
SELECT COUNT(*) FROM admin_media WHERE folder_id = $1;

-- name: CountAdminMediaUnfiled :one
SELECT COUNT(*) FROM admin_media WHERE folder_id IS NULL;

-- name: MoveAdminMediaToFolder :exec
UPDATE admin_media SET folder_id = $1, date_modified = $2 WHERE admin_media_id = $3;
