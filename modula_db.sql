PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL UNIQUE
);
INSERT INTO roles VALUES(1,'admin','{"admin":true,"create":true,"read":true,"update":true,"delete":true}');
INSERT INTO roles VALUES(2,'editor','{"admin":false,"create":true,"read":true,"update":true,"delete":false}');
INSERT INTO roles VALUES(4,'contributor','{"admin":false,"create":true,"read":true,"update":false,"delete":false}');
INSERT INTO roles VALUES(5,'subscriber','{"admin":false,"create":false,"read":true,"update":false,"delete":false}');

CREATE TABLE users (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users VALUES(1,'admin','System Administrator','admin@example.com','$2a$12$kZJ9.UG8cZ5oM4XB5IFk0uHGFCrL7dR.Cg9VeJDYQ/HU/V07yvVLq',1,'2025-03-11 15:57:55','2025-03-11 15:57:55');
INSERT INTO users VALUES(2,'demo','Demo User','demo@example.com','$2a$12$ZxJZLrfE7UGVlnHPrJsOR.k94Z0QG6TFnGXYSdwVcRtE8wE2NbL2y',2,'2025-03-11 15:57:55','2025-03-11 15:57:55');

CREATE TABLE sessions (
    session_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);

CREATE TABLE tokens (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
);

CREATE TABLE user_oauth (
    user_oauth_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    oauth_provider TEXT NOT NULL,
    oauth_provider_user_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(oauth_provider, oauth_provider_user_id)
);

CREATE TABLE routes (
    route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO routes VALUES(1,'/','Home',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO routes VALUES(2,'/about','About Us',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO routes VALUES(3,'/contact','Contact Us',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO routes VALUES(4,'/blog','Blog',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO routes VALUES(5,'/login','Login',1,'admin',1,'2025-03-14 14:05:18','2025-03-14 14:05:18',NULL);
INSERT INTO routes VALUES(6,'/register','Register',1,'admin',1,'2025-03-14 14:05:53','2025-03-14 14:05:53',NULL);

CREATE TABLE admin_routes (
    admin_route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author TEXT DEFAULT 'system',
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO admin_routes VALUES(1,'dashboard','Dashboard',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_routes VALUES(2,'users','User Management',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_routes VALUES(3,'content','Content Management',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_routes VALUES(4,'settings','System Settings',1,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);

CREATE TABLE tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
);
INSERT INTO tables VALUES(1,'admin_content_data',1);
INSERT INTO tables VALUES(2,'admin_content_fields',1);
INSERT INTO tables VALUES(3,'admin_datatypes',1);
INSERT INTO tables VALUES(4,'admin_fields',1);
INSERT INTO tables VALUES(5,'admin_routes',1);
INSERT INTO tables VALUES(6,'content_data',1);
INSERT INTO tables VALUES(7,'datatypes',1);
INSERT INTO tables VALUES(8,'fields',1);
INSERT INTO tables VALUES(9,'media',1);
INSERT INTO tables VALUES(10,'media_dimensions',1);
INSERT INTO tables VALUES(11,'roles',1);
INSERT INTO tables VALUES(12,'routes',1);
INSERT INTO tables VALUES(13,'sessions',1);
INSERT INTO tables VALUES(14,'tables',1);
INSERT INTO tables VALUES(15,'tokens',1);
INSERT INTO tables VALUES(16,'user_oauth',1);
INSERT INTO tables VALUES(17,'users',1);

CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL,
    datatype_id INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO content_data VALUES(1,1,1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_data VALUES(2,2,1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_data VALUES(3,4,2,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);

CREATE TABLE content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL,
    content_data_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    field_value TEXT NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO content_fields VALUES(1,1,1,1,'Welcome to ModulaCMS','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(2,1,1,2,'<h1>Welcome to ModulaCMS</h1><p>This is a flexible content management system that allows you to create and manage your website content easily.</p><p>Start by navigating to the admin panel to create your first post or page.</p>','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(3,2,2,1,'About Our Company','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(4,2,2,2,'<h1>About ModulaCMS</h1><p>ModulaCMS is a modern content management system built with Go and designed for flexibility and performance.</p><p>Our team is dedicated to creating tools that make web publishing accessible to everyone.</p>','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(5,4,3,1,'Getting Started with ModulaCMS','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(6,4,3,2,'<h1>Getting Started with ModulaCMS</h1><p>In this post, we will explore the basics of setting up and using ModulaCMS for your website or application.</p><p>ModulaCMS provides a powerful and flexible foundation for creating dynamic web content with minimal effort.</p>','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(7,4,3,4,'tutorial,beginners,setup','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO content_fields VALUES(8,4,3,5,'2023-01-15 10:30:00','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
CREATE TABLE admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER NOT NULL,
    admin_datatype_id INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO admin_content_data VALUES(1,1,4,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_content_data VALUES(2,1,0,'2025-03-17T17:07:14-04:00','2025-03-17T17:07:14-04:00',NULL);
INSERT INTO admin_content_data VALUES(3,5,0,'2025-03-17T17:09:04-04:00','2025-03-17T17:09:04-04:00',NULL);
INSERT INTO admin_content_data VALUES(4,4,0,'2025-03-17T17:10:46-04:00','2025-03-17T17:10:46-04:00',NULL);
INSERT INTO admin_content_data VALUES(5,2,0,'2025-03-18T05:35:17-04:00','2025-03-18T05:35:17-04:00',NULL);
INSERT INTO admin_content_data VALUES(6,2,5,'2025-03-18T05:35:34-04:00','2025-03-18T05:35:34-04:00',NULL);
INSERT INTO admin_content_data VALUES(7,2,5,'2025-03-18T05:35:17-04:00','2025-03-18T05:35:17-04:00',NULL);
CREATE TABLE admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER NOT NULL,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO admin_content_fields VALUES(1,1,1,5,'ModulaCMS Dashboard','2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
CREATE TABLE media_dimensions (
    md_id INTEGER PRIMARY KEY,
    label TEXT,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);
INSERT INTO media_dimensions VALUES(1,'Mobile',480,320,'3:2');
INSERT INTO media_dimensions VALUES(2,'Tablet',800,600,'4:3');
INSERT INTO media_dimensions VALUES(3,'Desktop',1920,1080,'16:9');
INSERT INTO media_dimensions VALUES(4,'UltraWide',3440,1440,'21:9');
CREATE TABLE media (
    media_id INTEGER PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO media VALUES(1,'sample-image.jpg','Sample Image','A sample image','Sample image caption','This is a sample image for demonstration purposes',NULL,'image/jpeg',NULL,'/media/sample-image.jpg',NULL,NULL,NULL,NULL,'admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55');

CREATE TABLE IF NOT EXISTS "admin_datatypes"
(
    admin_datatype_id INTEGER
        primary key,
    parent_id         INTEGER default NULL,
    label             TEXT                     not null,
    type              TEXT                     not null,
    author            TEXT    default 'system' not null,
    author_id         INTEGER default 1        not null,
    date_created      TEXT    default CURRENT_TIMESTAMP,
    date_modified     TEXT    default CURRENT_TIMESTAMP,
    history           TEXT
);
INSERT INTO admin_datatypes VALUES(1,NULL,'Page Editor','page-editor','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(2,NULL,'Post Editor','post-editor','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(3,NULL,'User Editor','user-editor','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(4,NULL,'Settings Editor','settings-editor','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(5,NULL,'Media Editor','media-editor','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(6,1,'Test','Page','admin ',1,'2025-03-11 15:57:55','2025-03-11 15:57:55','');
INSERT INTO admin_datatypes VALUES(7,0,'b','b','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55','');
INSERT INTO admin_datatypes VALUES(8,NULL,'c','c','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(9,NULL,'d','d','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(10,NULL,'e','e','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(11,NULL,'f','f','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(12,NULL,'g','g','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(13,NULL,'h','h','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(14,NULL,'j','j','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(15,NULL,'k','k','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(16,NULL,'l','l','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(17,NULL,'m','m','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(18,NULL,'n','n','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_datatypes VALUES(19,NULL,'o','o','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
CREATE TABLE IF NOT EXISTS "admin_fields"
(
    admin_field_id INTEGER
        primary key,
    parent_id      INTEGER default NULL,
    label          TEXT                     not null,
    data           TEXT,
    type           TEXT                     not null,
    author         TEXT    default 'system' not null,
    author_id      INTEGER default 1        not null,
    date_created   TEXT    default CURRENT_TIMESTAMP,
    date_modified  TEXT    default CURRENT_TIMESTAMP,
    history        TEXT
);
INSERT INTO admin_fields VALUES(1,NULL,'Username','{"required":true,"minLength":3,"maxLength":50}','text','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_fields VALUES(2,NULL,'Email','{"required":true,"format":"email"}','email','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_fields VALUES(3,NULL,'Password','{"required":true,"minLength":8}','password','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_fields VALUES(4,NULL,'Role','{"required":true,"options":"roles"}','select','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO admin_fields VALUES(5,NULL,'Site Title','{"required":true}','text','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
CREATE TABLE IF NOT EXISTS "datatypes"
(
    datatype_id   INTEGER
        primary key,
    parent_id     INTEGER default NULL,
    label         TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default 'system' not null,
    author_id     INTEGER default 1        not null,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP,
    history       TEXT
);
INSERT INTO datatypes VALUES(1,NULL,'Page','page','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO datatypes VALUES(2,NULL,'Post','post','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO datatypes VALUES(3,NULL,'Media','media','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO datatypes VALUES(4,NULL,'Widget','widget','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
CREATE TABLE IF NOT EXISTS "fields"
(
    field_id      INTEGER
        primary key,
    parent_id     INTEGER default NULL,
    label         TEXT                     not null,
    data          TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default 'system' not null,
    author_id     INTEGER default 1        not null,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP,
    history       TEXT
);
INSERT INTO fields VALUES(1,1,'Title','{"required":true,"default":""}','text','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO fields VALUES(2,1,'Content','{"required":true,"default":""}','richtext','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO fields VALUES(3,1,'Featured Image','{"required":false,"default":""}','image','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO fields VALUES(4,1,'Tags','{"required":false,"default":"","multiple":true}','tag','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO fields VALUES(5,1,'Publish Date','{"required":true,"default":"now"}','datetime','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
INSERT INTO fields VALUES(6,1,'Status','{"status":"published"}','status','admin',1,'2025-03-11 15:57:55','2025-03-11 15:57:55',NULL);
COMMIT;
