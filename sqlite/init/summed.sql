CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY,
    datecreated TEXT,
    datemodified TEXT,
    username TEXT,
    name TEXT,
    email TEXT UNIQUE,
    hash TEXT,
    role TEXT
);
CREATE TABLE IF NOT EXISTS adminroutes (
    author TEXT, 
    authorid TEXT, 
    slug TEXT UNIQUE, 
    title TEXT, 
    status INTEGER, 
    datecreated INTEGER, 
    datemodified INTEGER, 
    content TEXT, 
    template TEXT
);
CREATE TABLE IF NOT EXISTS routes (
    author TEXT, 
    authorid TEXT, 
    slug TEXT UNIQUE, 
    title TEXT, 
    status INTEGER, 
    datecreated INTEGER, 
    datemodified INTEGER, 
    content TEXT, 
    template TEXT
);
CREATE TABLE IF NOT EXISTS fields (
    id INTEGER PRIMARY KEY,
    routeid INTEGER,
    author TEXT,
    authorid TEXT,
    key TEXT,
    type TEXT,
    data TEXT,
    datecreated TEXT,
    datemodified TEXT,
    componentid INTEGER,
    tags TEXT,
    parent TEXT
);
CREATE TABLE IF NOT EXISTS elements(
    id INTEGER PRIMARY KEY,
    fieldid INTEGER,
    tag TEXT,
    FOREIGN KEY (fieldid) REFERENCES fields(id)
);
CREATE TABLE IF NOT EXISTS attributes (
    id INTEGER PRIMARY KEY,
    elementid INTEGER,
    key TEXT,
    value TEXT,
    FOREIGN KEY (elementid) REFERENCES elements(id)
);

CREATE TABLE IF NOT EXISTS media_dimensions (label TEXT UNIQUE, width INTEGER, height INTEGER);

CREATE TABLE IF NOT EXISTS media(
    name  TEXT,
    displayname TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT,
    authorid INTEGER,
    datecreated TEXT,
    datemodified TEXT,
    url TEXT UNIQUE,
    mimetype TEXT,
    dimensions TEXT,
    optimizedmobile TEXT,
    optimizedtablet TEXT,
    optimizeddesktop TEXT,
    optimizedultrawide TEXT
);

CREATE TABLE IF NOT EXISTS tables (id INTEGER PRIMARY KEY, label TEXT UNIQUE);



INSERT INTO tables (label) VALUES ('tables');
INSERT INTO tables (label) VALUES ('fields');
INSERT INTO tables (label) VALUES ('media');
INSERT INTO tables (label) VALUES ('routes');
INSERT INTO tables (label) VALUES ('adminroutes');
INSERT INTO tables (label) VALUES ('users');
INSERT INTO tables (label) VALUES ('elements');
INSERT INTO tables (label) VALUES ('attributes');

INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/','system',0,'home',0,'1111111111','1111111111','content','default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/pages','system',0 ,'pages', 0, '1111111111', '1111111111', 'content', 'default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/types','system',0 ,'types', 0, '1111111111', '1111111111', 'content', 'default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/fields','system',0, 'fields', 0, '1111111111', '1111111111', 'content', 'default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/menus','system',0 ,'menus', 0, '1111111111', '1111111111', 'content', 'default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/users','system',0, 'users', 0, '1111111111', '1111111111', 'content', 'default.html');
INSERT INTO adminroutes(slug,author,authorId,title,status,datecreated,datemodified,content,template)
VALUES ('/media','system',0, 'media', 0, '1111111111', '1111111111', 'content', 'default.html');
    
INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (1, 'Alice', 'alice123', 'field_key_1', 'text', 'Sample data for field 1', '%s', '%s', 1, 'tag1,tag2', 'parent1');
INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (2, 'Bob', 'bob456', 'field_key_2', 'number', '42', '%s', '%s', 2, 'tag3,tag4', 'parent2');
INSERT INTO fields (routeid, author, authorid, key, type, data, datecreated, datemodified, componentid, tags, parent)
VALUES (3, 'Charlie', 'charlie789', 'field_key_3', 'text', 'Another example of field data', '%s', '%s', 3, 'tag5', 'parent3');

INSERT INTO elements (id,fieldid, tag) VALUES
(1,1 ,'div');
INSERT INTO elements (id,fieldid, tag) VALUES
(2,1, 'span');
INSERT INTO elements (id,fieldid, tag) VALUES
(3,1 ,'section');
INSERT INTO routes (author, authorid, slug, title, status, datecreated, datemodified, content, template) VALUES 
('system','0','/place','Place','0','1111111111','1111111111','content','page.html');

INSERT INTO users (datecreated, datemodified, username, name, email, hash, role) VALUES 
('1111111111','1111111111','system', 'system', 'system@system.com', 'hash', 'root');


INSERT INTO attributes (elementid, key, value) VALUES
(1, 'class', 'container');
INSERT INTO attributes (elementid, key, value) VALUES
(1, 'id', 'main-div');
INSERT INTO attributes (elementid, key, value) VALUES
(2, 'style', 'color: red; font-size: 14px;');
INSERT INTO attributes (elementid, key, value) VALUES
(2, 'data-role', 'user-info');
INSERT INTO attributes (elementid, key, value) VALUES
(3, 'class', 'content-section');
INSERT INTO attributes (elementid, key, value) VALUES
(3, 'data-id', 'section-123');

