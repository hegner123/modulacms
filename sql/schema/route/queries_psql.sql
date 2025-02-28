-- name: GetRoute :one
SELECT * FROM routes
WHERE slug = $1
LIMIT 1;

-- name: CountRoute :one
SELECT COUNT(*)
FROM routes;

-- name: GetRouteId :one
SELECT route_id FROM routes
WHERE slug = $1
LIMIT 1;

-- name: ListRoute :many
SELECT * FROM routes
ORDER BY slug;

-- name: CreateRoute :one
INSERT INTO routes (
    author,
    author_id,
    slug,
    title,
    status,
    history,
    date_created,
    date_modified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateRoute :exec
UPDATE routes
SET slug = $1,
    title = $2,
    status = $3,
    history = $4,
    author = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8
WHERE slug = $9
RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE slug = $1;

