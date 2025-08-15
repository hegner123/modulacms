-- name: DropRouteTable :exec
DROP TABLE routes;

-- name: CreateRouteTable :exec
CREATE TABLE IF NOT EXISTS routes (
    route_id INT AUTO_INCREMENT
        PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    status INT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    history TEXT NULL,
    CONSTRAINT unique_slug
        UNIQUE (slug),
    CONSTRAINT fk_routes_routes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);

-- name: CountRoute :one
SELECT COUNT(*) 
FROM routes;

-- name: GetRoute :one
SELECT * FROM routes
WHERE route_id = ? 
LIMIT 1;

-- name: GetRouteIDBySlug :one
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
    slug,
    title,
    status,
    author_id,
    date_created,
    date_modified,
    history
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
 );

-- name: UpdateRoute :exec
UPDATE routes
SET slug = ?,
    title = ?,
    status = ?,
    history = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE slug = ?;

-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = ?;

