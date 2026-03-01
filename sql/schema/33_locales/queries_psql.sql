-- name: DropLocaleTable :exec
DROP TABLE locales;

-- name: CreateLocaleTable :exec
CREATE TABLE IF NOT EXISTS locales (
    locale_id     TEXT PRIMARY KEY NOT NULL,
    code          TEXT NOT NULL UNIQUE,
    label         TEXT NOT NULL,
    is_default    BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    fallback_code TEXT,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- name: CountLocale :one
SELECT COUNT(*)
FROM locales;

-- name: GetLocale :one
SELECT * FROM locales
WHERE locale_id = $1 LIMIT 1;

-- name: GetLocaleByCode :one
SELECT * FROM locales
WHERE code = $1 LIMIT 1;

-- name: GetDefaultLocale :one
SELECT * FROM locales
WHERE is_default = TRUE LIMIT 1;

-- name: ListLocales :many
SELECT * FROM locales
ORDER BY sort_order, code;

-- name: ListEnabledLocales :many
SELECT * FROM locales
WHERE is_enabled = TRUE
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
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING *;

-- name: UpdateLocale :exec
UPDATE locales
SET code = $1,
    label = $2,
    is_default = $3,
    is_enabled = $4,
    fallback_code = $5,
    sort_order = $6,
    date_created = $7
WHERE locale_id = $8;

-- name: DeleteLocale :exec
DELETE FROM locales
WHERE locale_id = $1;

-- name: ClearDefaultLocale :exec
UPDATE locales SET is_default = FALSE WHERE is_default = TRUE;

-- name: ListLocalesPaginated :many
SELECT * FROM locales
ORDER BY sort_order, code
LIMIT $1 OFFSET $2;
