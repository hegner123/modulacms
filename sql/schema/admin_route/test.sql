PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE users (
    user_id INTEGER
        PRIMARY KEY,

    username TEXT NOT NULL
        UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users VALUES(1,'system','system','system@modulacms.com','1920874301927','admin','12380192','9182370912');
INSERT INTO users VALUES(2,'user','user','user@agency.com','131923719237019','admin','12380192','9182370912');
INSERT INTO users VALUES(3,'wheel','wheel','wheel@agency.com','92740293874','editor','12380192','9182370912');
CREATE TABLE admin_routes (
    admin_route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    template TEXT DEFAULT "modula_base.html" NOT NULL
);
INSERT INTO admin_routes VALUES(1,'/','ModulaCms',0,'systm',1,'2024-12-03 20:48:17','2024-12-03 20:48:17','modula_base.html');
INSERT INTO admin_routes VALUES(2,'/admin/login','ModulaCMS',0,'system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17','modula_login.html');
CREATE TABLE tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);
INSERT INTO tables VALUES(1,'admin_datatypes',1);
INSERT INTO tables VALUES(2,'admin_fields',1);
INSERT INTO tables VALUES(3,'admin_routes',1);
INSERT INTO tables VALUES(4,'datatypes',1);
INSERT INTO tables VALUES(5,'fields',1);
INSERT INTO tables VALUES(6,'media',1);
INSERT INTO tables VALUES(7,'media_dimensions',1);
INSERT INTO tables VALUES(8,'tables',1);
INSERT INTO tables VALUES(9,'tokens',1);
INSERT INTO tables VALUES(10,'users',1);
CREATE TABLE routes (
    route_id INTEGER
        PRIMARY KEY,
    author TEXT DEFAULT "system" NOT NULL
        REFERENCES users (username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO routes VALUES(1,'system',1,'/','Home',0,'2024-12-03 20:48:17','2024-12-03 20:48:17');
CREATE TABLE tokens (
    id INTEGER
        PRIMARY KEY,
    user_id INTEGER NOT NULL
        REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL
        UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
);
INSERT INTO tokens VALUES(1,1,'Refresh','Test_Token','time','time+15',0);
INSERT INTO tokens VALUES(2,3,'Access','Test_token2','time','time+1',0);
CREATE TABLE IF NOT EXISTS "admin_datatypes"
(
    admin_dt_id    INTEGER
        primary key,
    admin_route_id INTEGER default NULL
        references admin_routes
            on update cascade on delete set default,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    not null,
    type           TEXT    not null,
    author         TEXT    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER not null
        references users
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP
);
INSERT INTO admin_datatypes VALUES(1,NULL,NULL,'GLOBALS','GLOBALS','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(2,NULL,1,'MENU','MENU','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(3,NULL,1,'SIDEBAR','SIDEBAR','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(4,NULL,2,'ADMIN_BAR_PRIMARY','ADMIN_BAR','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(5,NULL,2,'ADMIN_BAR_SECONDARY','ADMIN_BAR_SECONDARY','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(6,NULL,4,'PRIMARY_LINKS','PRIMARY_LINKS','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(7,NULL,6,'PLINK_1','PLINK_1','system',1,'9182370912','9182370912');
INSERT INTO admin_datatypes VALUES(8,NULL,6,'PLINK_2','PLINK_2','system',1,'9182370912','9182370912');
CREATE TABLE IF NOT EXISTS "admin_fields"
(
    admin_field_id INTEGER
        primary key,
    admin_route_id INTEGER default 1
        references admin_routes
            on update cascade on delete set default,
    parent_id      INTEGER default NULL
        references admin_datatypes
            on update cascade on delete set default,
    label          TEXT    default "unlabeled" not null,
    data           TEXT    default ""          not null,
    type           TEXT    default "text"      not null,
    author         TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id      INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP
);
INSERT INTO admin_fields VALUES(1,NULL,7,'Href','https://modulacms.com','url','system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17');
INSERT INTO admin_fields VALUES(2,NULL,7,'Text','Site','text','system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17');
INSERT INTO admin_fields VALUES(3,NULL,8,'Href','htts://modulacms.com','url','system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17');
CREATE TABLE IF NOT EXISTS "datatypes"
(
    datatype_id   INTEGER
        primary key,
    route_id      INTEGER default NULL
        references routes
            on update cascade on delete set default,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
);
INSERT INTO datatypes VALUES(1,1,NULL,'Home','Page','system',1,'time','time');
INSERT INTO datatypes VALUES(2,NULL,NULL,'Menu','Nav','system',1,'time','time');
INSERT INTO datatypes VALUES(3,NULL,NULL,'Footer','Section','system',1,'time','time');
INSERT INTO datatypes VALUES(4,NULL,NULL,'Socials','Widget','system',1,'time','time');
INSERT INTO datatypes VALUES(5,NULL,NULL,'Contact Us','Form','system',1,'time','time');
INSERT INTO datatypes VALUES(6,1,1,'Hero Header','Element','system',1,'time','time');
CREATE TABLE IF NOT EXISTS "fields"
(
    field_id      INTEGER
        primary key,
    route_id      INTEGER default NULL
        references routes
            on update cascade on delete set default,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT    default "unlabeled" not null,
    data          TEXT                        not null,
    type          TEXT                        not null,
    author        TEXT    default "system"    not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1           not null
        references users (user_id)
            on update cascade on delete set default,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
);
INSERT INTO fields VALUES(1,1,1,'url','{"media_id":"1"}','Image','system',1,'time','time');
INSERT INTO fields VALUES(2,1,1,'heading','Elevating fine dining to a microwave near you.','Heading','system',1,'time','time');
CREATE TABLE IF NOT EXISTS "media"
(
    media_id             INTEGER
        primary key,
    name                 TEXT,
    display_name         TEXT,
    alt                  TEXT,
    caption              TEXT,
    description          TEXT,
    class                TEXT,
    author               TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id            INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    date_created         TEXT    default CURRENT_TIMESTAMP,
    date_modified        TEXT    default CURRENT_TIMESTAMP,
    mimetype             TEXT,
    dimensions           TEXT,
    url                  TEXT
        unique,
    optimized_mobile     TEXT,
    optimized_tablet     TEXT,
    optimized_desktop    TEXT,
    optimized_ultra_wide TEXT
);
INSERT INTO media VALUES(1,'Test1.png','Test 1',NULL,NULL,NULL,NULL,'system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17','png',NULL,'https://placeholder.co?v=2',NULL,NULL,NULL,NULL);
INSERT INTO media VALUES(2,'Test2.png','Test 2',NULL,NULL,NULL,NULL,'system',1,'2024-12-03 20:48:17','2024-12-03 20:48:17','png',NULL,'https://placeholder.co',NULL,NULL,NULL,NULL);
CREATE TABLE IF NOT EXISTS "media_dimensions"
(
    md_id         INTEGER
        primary key,
    label         TEXT
        unique,
    width         INTEGER,
    height        INTEGER,
    aspect_ration TEXT
);
INSERT INTO media_dimensions VALUES(1,'mobile',300,600,'1:2');
INSERT INTO media_dimensions VALUES(2,'tablet',400,400,'1:1');
COMMIT;
