CREATE TABLE IF NOT EXISTS field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES admin_fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_modified    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
