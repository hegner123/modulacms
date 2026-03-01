CREATE TABLE IF NOT EXISTS locales (
    locale_id     TEXT PRIMARY KEY NOT NULL CHECK (length(locale_id) = 26),
    code          TEXT NOT NULL UNIQUE,
    label         TEXT NOT NULL,
    is_default    INTEGER NOT NULL DEFAULT 0,
    is_enabled    INTEGER NOT NULL DEFAULT 1,
    fallback_code TEXT,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    date_created  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_locales_code ON locales(code);
CREATE INDEX IF NOT EXISTS idx_locales_default ON locales(is_default) WHERE is_default = 1;
