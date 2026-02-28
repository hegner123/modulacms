-- name: CreatePluginsTable :exec
CREATE TABLE IF NOT EXISTS plugins (
    plugin_id TEXT PRIMARY KEY NOT NULL CHECK (length(plugin_id) = 26),
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'installed',
    capabilities JSONB NOT NULL DEFAULT '[]',
    approved_access JSONB NOT NULL DEFAULT '{}',
    manifest_hash TEXT NOT NULL DEFAULT '',
    date_installed TIMESTAMPTZ NOT NULL,
    date_modified TIMESTAMPTZ NOT NULL
);

-- name: CreatePluginsIndexStatus :exec
CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);

-- name: CreatePluginsIndexName :exec
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);

-- name: DropPluginsTable :exec
DROP TABLE IF EXISTS plugins;

-- name: GetPlugin :one
SELECT * FROM plugins WHERE plugin_id = $1 LIMIT 1;

-- name: GetPluginByName :one
SELECT * FROM plugins WHERE name = $1 LIMIT 1;

-- name: ListPlugins :many
SELECT * FROM plugins ORDER BY name;

-- name: ListPluginsByStatus :many
SELECT * FROM plugins WHERE status = $1 ORDER BY name;

-- name: CreatePlugin :one
INSERT INTO plugins (plugin_id, name, version, description, author, status, capabilities, approved_access, manifest_hash, date_installed, date_modified)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: UpdatePlugin :exec
UPDATE plugins SET version = $2, description = $3, author = $4, status = $5, capabilities = $6, approved_access = $7, manifest_hash = $8, date_modified = $9 WHERE plugin_id = $1;

-- name: UpdatePluginStatus :exec
UPDATE plugins SET status = $2, date_modified = $3 WHERE plugin_id = $1;

-- name: DeletePlugin :exec
DELETE FROM plugins WHERE plugin_id = $1;

-- name: CountPlugins :one
SELECT COUNT(*) FROM plugins;
