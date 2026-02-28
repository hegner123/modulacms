-- name: CreatePluginsTable :exec
CREATE TABLE IF NOT EXISTS plugins (
    plugin_id TEXT PRIMARY KEY NOT NULL CHECK (length(plugin_id) = 26),
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'installed',
    capabilities TEXT NOT NULL DEFAULT '[]',
    approved_access TEXT NOT NULL DEFAULT '{}',
    manifest_hash TEXT NOT NULL DEFAULT '',
    date_installed TEXT NOT NULL,
    date_modified TEXT NOT NULL
);

-- name: CreatePluginsIndexStatus :exec
CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);

-- name: CreatePluginsIndexName :exec
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);

-- name: DropPluginsTable :exec
DROP TABLE IF EXISTS plugins;

-- name: GetPlugin :one
SELECT * FROM plugins WHERE plugin_id = ? LIMIT 1;

-- name: GetPluginByName :one
SELECT * FROM plugins WHERE name = ? LIMIT 1;

-- name: ListPlugins :many
SELECT * FROM plugins ORDER BY name;

-- name: ListPluginsByStatus :many
SELECT * FROM plugins WHERE status = ? ORDER BY name;

-- name: CreatePlugin :one
INSERT INTO plugins (plugin_id, name, version, description, author, status, capabilities, approved_access, manifest_hash, date_installed, date_modified)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdatePlugin :exec
UPDATE plugins SET version = ?, description = ?, author = ?, status = ?, capabilities = ?, approved_access = ?, manifest_hash = ?, date_modified = ? WHERE plugin_id = ?;

-- name: UpdatePluginStatus :exec
UPDATE plugins SET status = ?, date_modified = ? WHERE plugin_id = ?;

-- name: DeletePlugin :exec
DELETE FROM plugins WHERE plugin_id = ?;

-- name: CountPlugins :one
SELECT COUNT(*) FROM plugins;
