
-- name: GetMedia :one
SELECT * FROM media
WHERE ? = ? LIMIT 1;

-- name: ListMedia :many
SELECT * FROM media
ORDER BY name;

-- name: CreateMedia :one
INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
 ?,?,? ,?,?,? ,?,?,? ,?,?,? ,?,?,? ,?,?
)
RETURNING *;

-- name: UpdateMedia :exec
UPDATE media
  set   name = ?,
        displayname = ?,
        alt = ?,
        caption = ?,
        description = ?,
        class = ?,
        author = ?,
        authorid = ?,
        datecreated = ?,
        datemodified = ?,
        url = ?,
        mimetype = ?,
        dimensions = ?,
        optimizedmobile = ?,
        optimizedtablet = ?,
        optimizeddesktop = ?,
        optimizedultrawide = ?
        WHERE ? = ?;

-- name: DeleteMedia :exec
DELETE FROM media
WHERE ? = ?;
