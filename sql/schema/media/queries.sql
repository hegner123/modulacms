
-- name: GetMedia :one
SELECT * FROM media
WHERE id = ? LIMIT 1;

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
    author,
    author_id,
    date_created,
    date_modified,
    url,
    mimetype,
    dimensions,
    optimized_mobile,
    optimized_tablet,
    optimized_desktop,
    optimized_ultrawide
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
        optimized_ultrawide = ?
        WHERE id = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE id = ?;
