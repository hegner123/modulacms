CREATE TABLE IF NOT EXISTS field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_field_plugin_config (
    field_id         TEXT PRIMARY KEY NOT NULL
        REFERENCES admin_fields(field_id) ON DELETE CASCADE,
    plugin_name      TEXT NOT NULL,
    plugin_interface TEXT NOT NULL,
    plugin_version   TEXT NOT NULL DEFAULT '',
    date_created     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
