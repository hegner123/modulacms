-- name: DropAdminMediaTable :exec
DROP TABLE admin_media;

-- name: CreateAdminMediaTable :exec
CREATE TABLE IF NOT EXISTS admin_media(
    admin_media_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_media_id) = 26),
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
    REFERENCES users
    ON DELETE SET NULL,
    folder_id TEXT NULL
    REFERENCES admin_media_folders(admin_folder_id)
    ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
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

-- name: CreateAdminMedia :one
INSERT INTO admin_media (
    admin_media_id,
    name,
    display_name,
    alt,
    caption,
    description,
    class,
    mimetype,
    dimensions,
    url,
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
)
RETURNING *;

-- name: UpdateAdminMedia :exec
UPDATE admin_media
SET name = ?,
    display_name = ?,
    alt = ?,
    caption = ?,
    description = ?,
    class = ?,
    mimetype = ?,
    dimensions = ?,
    url = ?,
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
