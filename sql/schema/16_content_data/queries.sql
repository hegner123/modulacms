-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    first_child_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    next_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    prev_sibling_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);



-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;

-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id;

-- name: CreateContentData :one
INSERT INTO content_data (
    route_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    datatype_id,
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
    ?
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = ?, 
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    datatype_id = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;
