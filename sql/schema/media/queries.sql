-- name: CreateMediaTable :exec
CREATE TABLE IF NOT EXISTS media
(
    media_id             INTEGER
        primary key,
    name                 TEXT,
    display_name         TEXT,
    alt                  TEXT,
    caption              TEXT,
    description          TEXT,
    class                TEXT,
    mimetype             TEXT,
    dimensions           TEXT,
    url                  TEXT
        unique,
    optimized_mobile     TEXT,
    optimized_tablet     TEXT,
    optimized_desktop    TEXT,
    optimized_ultra_wide TEXT,
    author               TEXT    default "system" not null
    references users (username)
    on update cascade on delete set default,
    author_id            INTEGER default 1        not null
    references users (user_id)
    on update cascade on delete set default,
    date_created         TEXT    default CURRENT_TIMESTAMP,
    date_modified        TEXT    default CURRENT_TIMESTAMP
);
-- name: GetMedia :one
SELECT * FROM media
WHERE media_id = ? LIMIT 1;

-- name: CountMedia :one
SELECT COUNT(*)
FROM media;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY name;

-- name: CreateMedia :one
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
    optimized_mobile,
    optimized_tablet,
    optimized_desktop,
    optimized_ultra_wide,
    author,
    author_id,
    date_created,
    date_modified
) VALUES (
 ?,?,? ,?,?,? ,?,?,? ,?,?,? ,?,?,? ,?,?
)
RETURNING *;

-- name: UpdateMedia :exec
UPDATE media
  set   name = ?,
        display_name = ?,
        alt = ?,
        caption = ?,
        description = ?,
        class = ?,
        author = ?,
        author_id = ?,
        date_created = ?,
        date_modified = ?,
        url = ?,
        mimetype = ?,
        dimensions = ?,
        optimized_mobile = ?,
        optimized_tablet = ?,
        optimized_desktop = ?,
        optimized_ultra_wide = ?
        WHERE media_id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE media_id = ?;
