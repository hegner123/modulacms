PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "media"
(
    media_id      INTEGER
        primary key,
    name          TEXT,
    display_name  TEXT,
    alt           TEXT,
    caption       TEXT,
    description   TEXT,
    class         TEXT,
    mimetype      TEXT,
    dimensions    TEXT,
    url           TEXT,
    srcset        TEXT,
    author        TEXT    default 'system' not null,
    author_id     INTEGER default 1        not null,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
);
INSERT INTO media VALUES(1,'background.jpeg','','','','','','','','','["https://modulacms.us-iad-10.linodeobjects.com/2025/3/background-1920x1080.jpeg","https://modulacms.us-iad-10.linodeobjects.com/2025/3/background-480x320.jpeg","https://modulacms.us-iad-10.linodeobjects.com/2025/3/background-800x600.jpeg","https://modulacms.us-iad-10.linodeobjects.com/2025/3/background-3440x1440.jpeg"]','admin',1,'','2025-03-20T09:10:09-04:00');
COMMIT;
