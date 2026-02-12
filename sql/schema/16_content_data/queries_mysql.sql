-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    first_child_id VARCHAR(26) NULL,
    next_sibling_id VARCHAR(26) NULL,
    prev_sibling_id VARCHAR(26) NULL,
    route_id VARCHAR(26) NULL,
    datatype_id VARCHAR(26) NULL,
    author_id VARCHAR(26) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_content_data_datatypes
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_first_child_id
        FOREIGN KEY (first_child_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_next_sibling_id
        FOREIGN KEY (next_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_prev_sibling_id
        FOREIGN KEY (prev_sibling_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
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

-- name: CreateContentData :exec
INSERT INTO content_data (
    content_data_id,
    route_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    datatype_id,
    author_id,
    status,
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
    ?
);

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = ?,
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListContentDataPaginated :many
SELECT * FROM content_data
ORDER BY content_data_id
LIMIT ? OFFSET ?;

-- name: ListContentDataByRoutePaginated :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id
LIMIT ? OFFSET ?;
