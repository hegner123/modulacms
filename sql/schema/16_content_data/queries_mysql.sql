-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE content_data (
    content_data_id INT AUTO_INCREMENT
        PRIMARY KEY,
    parent_id INT NULL,
    route_id INT NULL,
    datatype_id INT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    history TEXT NULL,
    CONSTRAINT fk_content_data_datatypes
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
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
    route_id,
    parent_id,
    datatype_id,
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

-- name: GetLastContentData :one
SELECT * FROM content_data WHERE content_data_id = LAST_INSERT_ID();

-- name: UpdateContentData :exec
UPDATE content_data
set route_id = ?,
    parent_id = ?,
    datatype_id = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;
