package main
//-- Blog-specific tables.
var termmetaTable = `CREATE TABLE termmeta (
    meta_id INTEGER PRIMARY KEY,
    term_id INTEGER NOT NULL DEFAULT 0,
    meta_key TEXT,
    meta_value TEXT
);`

var termsTable =`CREATE TABLE terms (
    term_id INTEGER PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    slug TEXT NOT NULL DEFAULT '',
    term_group INTEGER NOT NULL DEFAULT 0
);`

var termTaxonomyTable = `CREATE TABLE term_taxonomy (
    term_taxonomy_id INTEGER PRIMARY KEY,
    term_id INTEGER NOT NULL DEFAULT 0,
    taxonomy TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL,
    parent INTEGER NOT NULL DEFAULT 0,
    count INTEGER NOT NULL DEFAULT 0,
    UNIQUE (term_id, taxonomy)
);`

var termRelationshipsTable = `CREATE TABLE term_relationships (
    object_id INTEGER NOT NULL DEFAULT 0,
    term_taxonomy_id INTEGER NOT NULL DEFAULT 0,
    term_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (object_id, term_taxonomy_id)
);`

var commentmetaTable = `CREATE TABLE commentmeta (
    meta_id INTEGER PRIMARY KEY,
    comment_id INTEGER NOT NULL DEFAULT 0,
    meta_key TEXT,
    meta_value TEXT
);`

var commentsTable = `CREATE TABLE comments (
    comment_ID INTEGER PRIMARY KEY,
    comment_post_ID INTEGER NOT NULL DEFAULT 0,
    comment_author TEXT NOT NULL,
    comment_author_email TEXT NOT NULL DEFAULT '',
    comment_author_url TEXT NOT NULL DEFAULT '',
    comment_author_IP TEXT NOT NULL DEFAULT '',
    comment_date TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    comment_date_gmt TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    comment_content TEXT NOT NULL,
    comment_karma INTEGER NOT NULL DEFAULT 0,
    comment_approved TEXT NOT NULL DEFAULT '1',
    comment_agent TEXT NOT NULL DEFAULT '',
    comment_type TEXT NOT NULL DEFAULT 'comment',
    comment_parent INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL DEFAULT 0
);`

var linksTable =`CREATE TABLE links (
    link_id INTEGER PRIMARY KEY,
    link_url TEXT NOT NULL DEFAULT '',
    link_name TEXT NOT NULL DEFAULT '',
    link_image TEXT NOT NULL DEFAULT '',
    link_target TEXT NOT NULL DEFAULT '',
    link_description TEXT NOT NULL DEFAULT '',
    link_visible TEXT NOT NULL DEFAULT 'Y',
    link_owner INTEGER NOT NULL DEFAULT 1,
    link_rating INTEGER NOT NULL DEFAULT 0,
    link_updated TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    link_rel TEXT NOT NULL DEFAULT '',
    link_notes TEXT,
    link_rss TEXT NOT NULL DEFAULT ''
);`

var optionsTable = `CREATE TABLE options (
    option_id INTEGER PRIMARY KEY,
    option_name TEXT NOT NULL DEFAULT '',
    option_value TEXT NOT NULL,
    autoload TEXT NOT NULL DEFAULT 'yes',
    UNIQUE (option_name)
);`

var postmetaTable =`CREATE TABLE postmeta (
    meta_id INTEGER PRIMARY KEY,
    post_id INTEGER NOT NULL DEFAULT 0,
    meta_key TEXT,
    meta_value TEXT
);`

var postsTable =`CREATE TABLE posts (
    ID INTEGER PRIMARY KEY,
    post_author INTEGER NOT NULL DEFAULT 0,
    post_date TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    post_date_gmt TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    post_content TEXT NOT NULL,
    post_title TEXT NOT NULL,
    post_excerpt TEXT NOT NULL,
    post_status TEXT NOT NULL DEFAULT 'publish',
    comment_status TEXT NOT NULL DEFAULT 'open',
    ping_status TEXT NOT NULL DEFAULT 'open',
    post_password TEXT NOT NULL DEFAULT '',
    post_name TEXT NOT NULL DEFAULT '',
    to_ping TEXT NOT NULL,
    pinged TEXT NOT NULL,
    post_modified TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    post_modified_gmt TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    post_content_filtered TEXT NOT NULL,
    post_parent INTEGER NOT NULL DEFAULT 0,
    guid TEXT NOT NULL DEFAULT '',
    menu_order INTEGER NOT NULL DEFAULT 0,
    post_type TEXT NOT NULL DEFAULT 'post',
    post_mime_type TEXT NOT NULL DEFAULT '',
    comment_count INTEGER NOT NULL DEFAULT 0
);`

// Single site users table.
var usersTable =`CREATE TABLE users (
    ID INTEGER PRIMARY KEY,
    user_login TEXT NOT NULL DEFAULT '',
    user_pass TEXT NOT NULL DEFAULT '',
    user_nicename TEXT NOT NULL DEFAULT '',
    user_email TEXT NOT NULL DEFAULT '',
    user_url TEXT NOT NULL DEFAULT '',
    user_registered TEXT NOT NULL DEFAULT '0000-00-00 0000:00',
    user_activation_key TEXT NOT NULL DEFAULT '',
    user_status INTEGER NOT NULL DEFAULT 0,
    display_name TEXT NOT NULL DEFAULT ''
);`


//Usermeta table.
var userMetaTable = `CREATE TABLE usermeta (
    umeta_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL DEFAULT 0,
    meta_key TEXT,
    meta_value TEXT
);`

