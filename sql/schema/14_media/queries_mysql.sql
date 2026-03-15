-- name: DropMediaTable :exec
DROP TABLE media;

-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media (
    media_id VARCHAR(26) PRIMARY KEY NOT NULL,
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
    CONSTRAINT url
        UNIQUE (url),
    CONSTRAINT fk_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_media_media_folders_folder_id
        FOREIGN KEY (folder_id) REFERENCES media_folders (folder_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);


-- name: CountMedia :one
SELECT COUNT(*)
FROM media;

-- name: GetMedia :one
SELECT * FROM media
WHERE media_id = ? LIMIT 1;

-- name: GetMediaByName :one
SELECT * FROM media
WHERE name = ? LIMIT 1;

-- name: GetMediaByUrl :one
SELECT * FROM media
WHERE url = ? LIMIT 1;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY name;

-- name: CreateMedia :exec
INSERT INTO media (
    media_id,
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
-- name: UpdateMedia :exec
UPDATE media
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
WHERE media_id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = ?;

-- name: ListMediaPaginated :many
SELECT * FROM media
ORDER BY name
LIMIT ? OFFSET ?;

-- name: ListMediaByFolder :many
SELECT * FROM media WHERE folder_id = ? ORDER BY date_created DESC;

-- name: ListMediaByFolderPaginated :many
SELECT * FROM media WHERE folder_id = ? ORDER BY date_created DESC LIMIT ? OFFSET ?;

-- name: ListMediaUnfiled :many
SELECT * FROM media WHERE folder_id IS NULL ORDER BY date_created DESC;

-- name: ListMediaUnfiledPaginated :many
SELECT * FROM media WHERE folder_id IS NULL ORDER BY date_created DESC LIMIT ? OFFSET ?;

-- name: CountMediaByFolder :one
SELECT COUNT(*) FROM media WHERE folder_id = ?;

-- name: CountMediaUnfiled :one
SELECT COUNT(*) FROM media WHERE folder_id IS NULL;

-- name: MoveMediaToFolder :exec
UPDATE media SET folder_id = ?, date_modified = ? WHERE media_id = ?;
