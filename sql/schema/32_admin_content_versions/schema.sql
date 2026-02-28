CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_version_id) = 26),
    admin_content_data_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    snapshot TEXT NOT NULL,
    trigger TEXT NOT NULL DEFAULT 'manual',
    label TEXT NOT NULL DEFAULT '',
    published INTEGER NOT NULL DEFAULT 0,
    published_by TEXT
        REFERENCES users(user_id)
            ON DELETE SET NULL,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_acv_content ON admin_content_versions(admin_content_data_id);
CREATE INDEX IF NOT EXISTS idx_acv_content_locale ON admin_content_versions(admin_content_data_id, locale);
CREATE INDEX IF NOT EXISTS idx_acv_published ON admin_content_versions(admin_content_data_id, locale) WHERE published = 1;
