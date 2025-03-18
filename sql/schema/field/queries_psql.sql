-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id SERIAL PRIMARY KEY,
    
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT 'system',
    author_id INTEGER NOT NULL DEFAULT 1,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT,
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
$1,$2,$3,$4,$5,$6,$7,$8,$9
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
set parent_id = $1,
    label = $2,
    data = $3,
    type = $4,
    author = $5,
    author_id = $6,
    history =$7,
    date_created = $8,
    date_modified = $9
    WHERE field_id = $10
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = $1;

