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

CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);
CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);
