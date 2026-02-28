-- name: CreatePipelinesTable :exec
CREATE TABLE IF NOT EXISTS pipelines (
    pipeline_id TEXT PRIMARY KEY NOT NULL CHECK (length(pipeline_id) = 26),
    plugin_id TEXT NOT NULL REFERENCES plugins(plugin_id) ON UPDATE CASCADE ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    plugin_name TEXT NOT NULL,
    handler TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 50,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB NOT NULL DEFAULT '{}',
    date_created TIMESTAMPTZ NOT NULL,
    date_modified TIMESTAMPTZ NOT NULL
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
SELECT * FROM pipelines WHERE pipeline_id = $1 LIMIT 1;

-- name: ListPipelines :many
SELECT * FROM pipelines ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTable :many
SELECT * FROM pipelines WHERE table_name = $1 ORDER BY operation, priority;

-- name: ListPipelinesByPluginID :many
SELECT * FROM pipelines WHERE plugin_id = $1 ORDER BY table_name, operation, priority;

-- name: ListPipelinesByTableOperation :many
SELECT * FROM pipelines WHERE table_name = $1 AND operation = $2 ORDER BY priority, plugin_name;

-- name: ListEnabledPipelines :many
SELECT * FROM pipelines WHERE enabled = true ORDER BY table_name, operation, priority;

-- name: CreatePipeline :one
INSERT INTO pipelines (pipeline_id, plugin_id, table_name, operation, plugin_name, handler, priority, enabled, config, date_created, date_modified)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING *;

-- name: UpdatePipeline :exec
UPDATE pipelines SET table_name = $2, operation = $3, handler = $4, priority = $5, enabled = $6, config = $7, date_modified = $8 WHERE pipeline_id = $1;

-- name: UpdatePipelineEnabled :exec
UPDATE pipelines SET enabled = $2, date_modified = $3 WHERE pipeline_id = $1;

-- name: DeletePipeline :exec
DELETE FROM pipelines WHERE pipeline_id = $1;

-- name: DeletePipelinesByPluginID :exec
DELETE FROM pipelines WHERE plugin_id = $1;

-- name: CountPipelines :one
SELECT COUNT(*) FROM pipelines;
