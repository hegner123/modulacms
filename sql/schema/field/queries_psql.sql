-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id SERIAL PRIMARY KEY,
    route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT 'system',
    author_id INTEGER NOT NULL DEFAULT 1,
    history TEXT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_routes FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_datatypes FOREIGN KEY (parent_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author FOREIGN KEY (author)
        REFERENCES users(username)
        ON UPDATE CASCADE ON DELETE SET DEFAULT,
    CONSTRAINT fk_users_author_id FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET DEFAULT
);

-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = $1 LIMIT 1;

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: CreateField :one
INSERT INTO fields  (
    route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    history,
    date_created,
    date_modified
    ) VALUES (
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
set route_id = $1,
    parent_id = $2,
    label = $3,
    data = $4,
    type = $5,
    author = $6,
    author_id = $7,
    history =$8,
    date_created = $9,
    date_modified = $10
    WHERE field_id = $11
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = $1;

-- name: ListFieldByRouteId :many
SELECT field_id, route_id, parent_id, label, data, type
FROM fields 
WHERE route_id = $1;
