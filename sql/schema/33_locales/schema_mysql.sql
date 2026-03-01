CREATE TABLE IF NOT EXISTS locales (
    locale_id     VARCHAR(26) PRIMARY KEY NOT NULL,
    code          VARCHAR(35) NOT NULL UNIQUE,
    label         VARCHAR(255) NOT NULL,
    is_default    TINYINT NOT NULL DEFAULT 0,
    is_enabled    TINYINT NOT NULL DEFAULT 1,
    fallback_code VARCHAR(35),
    sort_order    INT NOT NULL DEFAULT 0,
    date_created  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_locales_code ON locales(code);
