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

CREATE INDEX IF NOT EXISTS idx_locales_code ON locales(code);
