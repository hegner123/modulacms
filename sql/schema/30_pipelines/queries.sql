-- name: CreatePipelinesTable :exec
CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id TEXT PRIMARY KEY NOT NULL CHECK (length(pipeline_id) = 26),
    plugin_id TEXT NOT NULL,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    plugin_name TEXT NOT NULL,
    handler TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    enabled INTEGER NOT NULL DEFAULT 1,
    config TEXT NOT NULL DEFAULT '{}',
    date_created TEXT NOT NULL,
    date_modified TEXT NOT NULL,
    FOREIGN KEY (plugin_id) REFERENCES plugins(plugin_id) ON DELETE CASCADE
);

-- name: CreatePipelinesIndexUnique :exec
CREATE UNIQUE INDEX IF NOT EXISTS idx_pipeline_unique ON pipelines(table_name, operation, plugin_id);

-- name: CreatePipelinesIndexPlugin :exec
CREATE INDEX IF NOT EXISTS idx_pipelines_plugin ON pipelines(plugin_id);

-- name: CreatePipelinesIndexTable :exec
CREATE INDEX IF NOT EXISTS idx_pipelines_table ON pipelines(table_name);

-- name: DropPipelinesTable :exec
DROP TABLE IF EXISTS pipelines;

-- name: GetPipeline :one
SELECT * FROM pipelines WHERE pipeline_id = ? LIMIT 1;

-- name: ListPipelines :many
SELECT * FROM pipelines ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTable :many
SELECT * FROM pipelines WHERE table_name = ? ORDER BY operation, priority;

-- name: ListPipelinesByPluginID :many
SELECT * FROM pipelines WHERE plugin_id = ? ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTableOperation :many
SELECT * FROM pipelines WHERE table_name = ? AND operation = ? ORDER BY priority, plugin_name;

-- name: ListEnabledPipelines :many
SELECT * FROM pipelines WHERE enabled = 1 ORDER BY table_name, operation, priority;

-- name: CreatePipeline :one
INSERT INTO pipelines (pipeline_id, plugin_id, table_name, operation, plugin_name, handler, priority, enabled, config, date_created, date_modified)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdatePipeline :exec
UPDATE pipelines SET table_name = ?, operation = ?, handler = ?, priority = ?, enabled = ?, config = ?, date_modified = ? WHERE pipeline_id = ?;

-- name: UpdatePipelineEnabled :exec
UPDATE pipelines SET enabled = ?, date_modified = ? WHERE pipeline_id = ?;

-- name: DeletePipeline :exec
DELETE FROM pipelines WHERE pipeline_id = ?;

-- name: DeletePipelinesByPluginID :exec
DELETE FROM pipelines WHERE plugin_id = ?;

-- name: CountPipelines :one
SELECT COUNT(*) FROM pipelines;
