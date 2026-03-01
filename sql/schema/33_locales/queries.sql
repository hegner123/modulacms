-- name: DropLocaleTable :exec
DROP TABLE locales;

-- name: CreateLocaleTable :exec
CREATE TABLE IF NOT EXISTS locales (
    locale_id     TEXT PRIMARY KEY NOT NULL CHECK (length(locale_id) = 26),
    code          TEXT NOT NULL UNIQUE,
    label         TEXT NOT NULL,
    is_default    INTEGER NOT NULL DEFAULT 0,
    is_enabled    INTEGER NOT NULL DEFAULT 1,
    fallback_code TEXT,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- name: CountLocale :one
SELECT COUNT(*)
FROM locales;

-- name: GetLocale :one
SELECT * FROM locales
WHERE locale_id = ? LIMIT 1;

-- name: GetLocaleByCode :one
SELECT * FROM locales
WHERE code = ? LIMIT 1;

-- name: GetDefaultLocale :one
SELECT * FROM locales
WHERE is_default = 1 LIMIT 1;

-- name: ListLocales :many
SELECT * FROM locales
ORDER BY sort_order, code;

-- name: ListEnabledLocales :many
SELECT * FROM locales
WHERE is_enabled = 1
ORDER BY sort_order, code;

-- name: CreateLocale :one
INSERT INTO locales (
    locale_id,
    code,
    label,
    is_default,
    is_enabled,
    fallback_code,
    sort_order,
    date_created
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: UpdateLocale :exec
UPDATE locales
SET code = ?,
    label = ?,
    is_default = ?,
    is_enabled = ?,
    fallback_code = ?,
    sort_order = ?,
    date_created = ?
WHERE locale_id = ?;

-- name: DeleteLocale :exec
DELETE FROM locales
WHERE locale_id = ?;

-- name: ClearDefaultLocale :exec
UPDATE locales SET is_default = 0 WHERE is_default = 1;

-- name: ListLocalesPaginated :many
SELECT * FROM locales
ORDER BY sort_order, code
LIMIT ? OFFSET ?;
