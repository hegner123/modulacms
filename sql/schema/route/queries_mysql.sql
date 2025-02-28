-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = ? 
LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*) 
FROM routes;

-- name: GetRouteId :one
SELECT route_id 
FROM routes
WHERE slug = ? 
LIMIT 1;

-- name: GetLastRoute :one
SELECT * FROM routes WHERE route_id = LAST_INSERT_ID();

-- name: ListRoute :many
SELECT * FROM routes
ORDER BY slug;

-- name: CreateRoute :exec
INSERT INTO routes (
    author,
    author_id,
    slug,
    title,
    status,
    history,
    date_created,
    date_modified
) VALUES (?,?,?,?,?,?,?,?);

-- name: UpdateRoute :exec
UPDATE routes
SET slug = ?,
    title = ?,
    status = ?,
    history = ?,
    author = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE slug = ?;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = ?;

