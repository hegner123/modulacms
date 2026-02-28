CREATE TABLE IF NOT EXISTS content_versions (
    content_version_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_version_id) = 26),
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
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

CREATE INDEX IF NOT EXISTS idx_cv_content ON content_versions(content_data_id);
CREATE INDEX IF NOT EXISTS idx_cv_content_locale ON content_versions(content_data_id, locale);
CREATE INDEX IF NOT EXISTS idx_cv_published ON content_versions(content_data_id, locale) WHERE published = 1;
