-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    route_id    INT DEFAULT NULL,
    parent_id    INT DEFAULT NULL,
    datatype_id INT DEFAULT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL,
    CONSTRAINT fk_content_data_route_id FOREIGN KEY (route_id)
        REFERENCES routes(route_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_parent_id FOREIGN KEY (parent_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_datatypes FOREIGN KEY (datatype_id)
        REFERENCES datatypes(datatype_id)
        ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;

-- name: CreateContentData :exec
INSERT INTO content_data (
    route_id,
    parent_id,
    datatype_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?,?,?
    );
-- name: GetLastContentData :one
SELECT * FROM content_data WHERE content_data_id = LAST_INSERT_ID();


-- name: UpdateContentData :exec
UPDATE content_data
set route_id = ?,
    parent_id = ?,
    datatype_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = ?;
