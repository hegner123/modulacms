-- name: DropMediaTable :exec
DROP TABLE media;

-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media (
    media_id INT AUTO_INCREMENT
        PRIMARY KEY,
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
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT url
        UNIQUE (url),
    CONSTRAINT fk_media_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
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
    author_id,
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
    ?
);
-- name: GetLastMedia :one
SELECT * FROM media WHERE media_id = LAST_INSERT_ID();

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
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE media_id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = ?;
