CREATE TABLE IF NOT EXISTS admin_content_versions (
    admin_content_version_id VARCHAR(26) NOT NULL,
    admin_content_data_id VARCHAR(26) NOT NULL,
    version_number INT NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT '',
    snapshot MEDIUMTEXT NOT NULL,
    `trigger` VARCHAR(50) NOT NULL DEFAULT 'manual',
    label VARCHAR(255) NOT NULL DEFAULT '',
    published TINYINT NOT NULL DEFAULT 0,
    published_by VARCHAR(26),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_content_version_id),
    CONSTRAINT fk_acv_content FOREIGN KEY (admin_content_data_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_acv_published_by FOREIGN KEY (published_by)
        REFERENCES users(user_id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX idx_acv_content ON admin_content_versions(admin_content_data_id);
CREATE INDEX idx_acv_content_locale ON admin_content_versions(admin_content_data_id, locale);
CREATE INDEX idx_acv_published ON admin_content_versions(admin_content_data_id, locale, published);
