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
    role INTEGER NOT NULL DEFAULT 4,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users VALUES(1,'admin','admin','admin@modulacms.com','saf',1,'2025-03-30 15:08:40','2025-03-30 15:08:40');
CREATE TABLE routes (
    route_id INTEGER
        PRIMARY KEY,
    slug TEXT NOT NULL
        UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
    REFERENCES users
    ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO routes VALUES(1,'/','home',1,1,'2025-04-01 11:00:59','2025-04-01 11:00:59',NULL);
CREATE TABLE datatypes(
    datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO datatypes VALUES(1,NULL,'Page','ROOT',1,'2025-04-01 11:25:42','2025-04-07T06:58:22-04:00',NULL);
INSERT INTO datatypes VALUES(2,NULL,'Post','ROOT',1,'2025-04-01 11:25:42','2025-04-07T06:58:27-04:00',NULL);
INSERT INTO datatypes VALUES(3,NULL,'Section','ROOT',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(4,1,'Hero','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(5,1,'Navigation','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(6,1,'Footer','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(7,1,'Container','Layout',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(8,7,'Row','Layout',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(9,8,'Column','Layout',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(10,9,'RichText','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(11,9,'Image','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(12,9,'Button','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(13,9,'Card','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(14,13,'CardHeader','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(15,13,'CardBody','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(16,13,'CardFooter','Content',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(17,3,'Feature','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(18,3,'Testimonial','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
INSERT INTO datatypes VALUES(19,3,'Gallery','Block',1,'2025-04-01 11:25:42','2025-04-01 11:25:42',NULL);
CREATE TABLE fields(
    field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO fields VALUES(1,1,'Title','{"required":true,"validation":".*(255)","placeholder":"Page title"}','Meta',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(2,1,'Favicon','{"required":false,"validation":"media","accept":"image/x-icon,image/png"}','Meta',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(3,1,'MetaDescription','{"required":false,"maxLength":160,"placeholder":"Brief description for search engines"}','Meta',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(4,1,'Keywords','{"required":false,"placeholder":"SEO keywords, comma separated"}','Meta',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(5,1,'OgImage','{"required":false,"validation":"media","accept":"image/*"}','Meta',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(6,4,'Headline','{"required":true,"maxLength":100,"placeholder":"Main hero headline"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(7,4,'Subheadline','{"required":false,"maxLength":200,"placeholder":"Supporting text"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(8,4,'BackgroundImage','{"required":false,"validation":"media","accept":"image/*"}','Media',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(9,4,'CtaText','{"required":false,"maxLength":50,"placeholder":"Call to action text"}','Text',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(10,4,'CtaUrl','{"required":false,"validation":"url","placeholder":"https://example.com"}','Url',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(11,5,'BrandName','{"required":true,"maxLength":50,"placeholder":"Company name"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(12,5,'Logo','{"required":false,"validation":"media","accept":"image/*"}','Media',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(13,5,'MenuItems','{"required":false,"type":"json","placeholder":"Navigation menu structure"}','Json',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(14,6,'CopyrightText','{"required":false,"placeholder":"© 2025 Company Name"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(15,6,'SocialLinks','{"required":false,"type":"json","placeholder":"Social media links"}','Json',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(16,6,'ContactInfo','{"required":false,"type":"json","placeholder":"Contact information"}','Json',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(17,7,'CssClass','{"required":false,"placeholder":"container, container-fluid, etc."}','Text',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(18,7,'MaxWidth','{"required":false,"placeholder":"1200px, 100%, etc."}','Text',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(19,8,'CssClass','{"required":false,"placeholder":"row, row-no-gutters, etc."}','Text',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(20,8,'JustifyContent','{"required":false,"options":["start","center","end","between","around"]}','Select',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(21,9,'CssClass','{"required":false,"placeholder":"col-12, col-md-6, etc."}','Text',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(22,9,'Order','{"required":false,"min":1,"max":12}','Number',1,'2025-04-08 12:35:53','2025-04-08 12:35:53',NULL);
INSERT INTO fields VALUES(23,10,'Content','{"required":true,"toolbar":"full","placeholder":"Rich text content..."}','RichText',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(24,11,'Src','{"required":true,"validation":"media","accept":"image/*"}','Media',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(25,11,'Alt','{"required":true,"maxLength":255,"placeholder":"Describe the image"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(26,11,'Caption','{"required":false,"maxLength":500,"placeholder":"Image caption"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(27,11,'Width','{"required":false,"placeholder":"100%, 300px, etc."}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(28,12,'Text','{"required":true,"maxLength":100,"placeholder":"Button text"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(29,12,'Url','{"required":true,"validation":"url","placeholder":"https://example.com"}','Url',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(30,12,'Style','{"required":false,"options":["primary","secondary","success","danger","warning","info"]}','Select',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(31,12,'Size','{"required":false,"options":["sm","md","lg"]}','Select',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(32,13,'CssClass','{"required":false,"placeholder":"card, card-shadow, etc."}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(33,14,'Title','{"required":false,"maxLength":100,"placeholder":"Card title"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(34,14,'Image','{"required":false,"validation":"media","accept":"image/*"}','Media',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(35,15,'Content','{"required":true,"toolbar":"basic","placeholder":"Card body content"}','RichText',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(36,16,'Links','{"required":false,"type":"json","placeholder":"Footer links for card"}','Json',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(37,17,'Title','{"required":true,"maxLength":100,"placeholder":"Feature title"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(38,17,'Description','{"required":true,"maxLength":500,"placeholder":"Feature description"}','Textarea',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(39,17,'Icon','{"required":false,"placeholder":"fa-star, fa-heart, etc."}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(40,18,'Quote','{"required":true,"maxLength":500,"placeholder":"Customer testimonial"}','Textarea',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(41,18,'Author','{"required":true,"maxLength":100,"placeholder":"Customer name"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(42,18,'Position','{"required":false,"maxLength":100,"placeholder":"Job title"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(43,18,'Company','{"required":false,"maxLength":100,"placeholder":"Company name"}','Text',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(44,18,'Avatar','{"required":false,"validation":"media","accept":"image/*"}','Media',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(45,19,'Images','{"required":true,"validation":"media","accept":"image/*","multiple":true}','Media',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(46,19,'Columns','{"required":false,"options":["2","3","4","6"],"default":"3"}','Select',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
INSERT INTO fields VALUES(47,19,'ShowCaptions','{"required":false,"default":false}','Boolean',1,'2025-04-01 11:33:47','2025-04-01 11:33:47',NULL);
CREATE TABLE content_data (
    content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL
);
INSERT INTO content_data VALUES(1,NULL,1,1,1,'2025-04-01 11:34:44','2025-04-04T10:12:19-04:00','');
INSERT INTO content_data VALUES(2,1,1,5,1,'2025-04-08 12:01:00','2025-04-08 12:01:00',NULL);
INSERT INTO content_data VALUES(3,1,1,4,1,'2025-04-08 12:02:50','2025-04-08 12:02:50',NULL);
INSERT INTO content_data VALUES(4,1,1,7,1,'2025-04-08 12:03:00','2025-04-08 12:03:00',NULL);
INSERT INTO content_data VALUES(5,4,1,8,1,'2025-04-08 12:33:52','2025-04-08 12:33:52',NULL);
INSERT INTO content_data VALUES(6,5,1,9,1,'2025-04-08 12:34:19','2025-04-08 12:34:19',NULL);
INSERT INTO content_data VALUES(7,6,1,10,1,'2025-04-08 12:34:30','2025-04-08 12:34:30',NULL);
INSERT INTO content_data VALUES(8,5,1,9,1,'2025-04-08 12:34:40','2025-04-08 12:34:40',NULL);
INSERT INTO content_data VALUES(9,8,1,13,1,'2025-04-08 12:35:00','2025-04-08 12:35:00',NULL);
INSERT INTO content_data VALUES(10,9,1,14,1,'2025-04-08 12:35:10','2025-04-08 12:35:10',NULL);
INSERT INTO content_data VALUES(11,9,1,15,1,'2025-04-08 12:35:20','2025-04-08 12:35:20',NULL);
INSERT INTO content_data VALUES(12,9,1,16,1,'2025-04-08 12:35:30','2025-04-08 12:35:30',NULL);
INSERT INTO content_data VALUES(13,5,1,9,1,'2025-04-08 12:36:00','2025-04-08 12:36:00',NULL);
INSERT INTO content_data VALUES(14,13,1,13,1,'2025-04-08 12:36:10','2025-04-08 12:36:10',NULL);
INSERT INTO content_data VALUES(15,14,1,14,1,'2025-04-08 12:36:20','2025-04-08 12:36:20',NULL);
INSERT INTO content_data VALUES(16,14,1,15,1,'2025-04-08 12:36:30','2025-04-08 12:36:30',NULL);
INSERT INTO content_data VALUES(17,14,1,16,1,'2025-04-08 12:36:40','2025-04-08 12:36:40',NULL);
INSERT INTO content_data VALUES(18,1,1,7,1,'2025-04-08 12:37:00','2025-04-08 12:37:00',NULL);
INSERT INTO content_data VALUES(19,18,1,8,1,'2025-04-08 12:37:10','2025-04-08 12:37:10',NULL);
INSERT INTO content_data VALUES(20,19,1,9,1,'2025-04-08 12:37:20','2025-04-08 12:37:20',NULL);
INSERT INTO content_data VALUES(21,20,1,17,1,'2025-04-08 12:37:30','2025-04-08 12:37:30',NULL);
INSERT INTO content_data VALUES(22,19,1,9,1,'2025-04-08 12:37:40','2025-04-08 12:37:40',NULL);
INSERT INTO content_data VALUES(23,22,1,17,1,'2025-04-08 12:37:50','2025-04-08 12:37:50',NULL);
INSERT INTO content_data VALUES(24,19,1,9,1,'2025-04-08 12:38:00','2025-04-08 12:38:00',NULL);
INSERT INTO content_data VALUES(25,24,1,17,1,'2025-04-08 12:38:10','2025-04-08 12:38:10',NULL);
INSERT INTO content_data VALUES(26,1,1,7,1,'2025-04-08 12:39:00','2025-04-08 12:39:00',NULL);
INSERT INTO content_data VALUES(27,26,1,8,1,'2025-04-08 12:39:10','2025-04-08 12:39:10',NULL);
INSERT INTO content_data VALUES(28,27,1,9,1,'2025-04-08 12:39:20','2025-04-08 12:39:20',NULL);
INSERT INTO content_data VALUES(29,28,1,18,1,'2025-04-08 12:39:30','2025-04-08 12:39:30',NULL);
INSERT INTO content_data VALUES(30,1,1,6,1,'2025-04-08 12:40:00','2025-04-08 12:40:00',NULL);
CREATE TABLE content_fields (
    content_field_id INTEGER
        PRIMARY KEY,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
INSERT INTO content_fields VALUES(1,1,1,1,'Quantum Widget Solutions - Making Tomorrow Today',1,'2025-04-01 11:35:10','2025-04-01 11:35:10',NULL);
INSERT INTO content_fields VALUES(2,1,1,3,'Discover revolutionary widgets that defy physics and common sense. Perfect for your inexplicable business needs.',1,'2025-04-01 11:35:10','2025-04-01 11:35:10',NULL);
INSERT INTO content_fields VALUES(3,1,1,4,'quantum widgets, interdimensional solutions, temporal commerce, widget technology, impossible business',1,'2025-04-01 11:35:10','2025-04-01 11:35:10',NULL);
INSERT INTO content_fields VALUES(4,1,2,11,'Quantum Widget Solutions',1,'2025-04-08 12:01:00','2025-04-08 12:01:00',NULL);
INSERT INTO content_fields VALUES(5,1,2,13,'[{"label":"Home","url":"/"},{"label":"Products","url":"/products"},{"label":"About","url":"/about"},{"label":"Contact","url":"/contact"}]',1,'2025-04-08 12:01:00','2025-04-08 12:01:00',NULL);
INSERT INTO content_fields VALUES(6,1,3,6,'Widgets So Advanced, They Confuse Themselves',1,'2025-04-08 12:02:50','2025-04-08 12:02:50',NULL);
INSERT INTO content_fields VALUES(7,1,3,7,'Experience the future of widget technology that makes absolutely no sense but works perfectly anyway.',1,'2025-04-08 12:02:50','2025-04-08 12:02:50',NULL);
INSERT INTO content_fields VALUES(8,1,3,9,'Explore Our Paradoxical Catalog',1,'2025-04-08 12:02:50','2025-04-08 12:02:50',NULL);
INSERT INTO content_fields VALUES(9,1,3,10,'/products',1,'2025-04-08 12:02:50','2025-04-08 12:02:50',NULL);
INSERT INTO content_fields VALUES(10,1,4,17,'container-fluid',1,'2025-04-08 12:03:00','2025-04-08 12:03:00',NULL);
INSERT INTO content_fields VALUES(11,1,4,18,'1200px',1,'2025-04-08 12:03:00','2025-04-08 12:03:00',NULL);
INSERT INTO content_fields VALUES(12,1,5,19,'row justify-content-center',1,'2025-04-08 12:33:52','2025-04-08 12:33:52',NULL);
INSERT INTO content_fields VALUES(13,1,5,20,'center',1,'2025-04-08 12:33:52','2025-04-08 12:33:52',NULL);
INSERT INTO content_fields VALUES(14,1,6,21,'col-md-4',1,'2025-04-08 12:34:19','2025-04-08 12:34:19',NULL);
INSERT INTO content_fields VALUES(15,1,6,22,'1',1,'2025-04-08 12:34:19','2025-04-08 12:34:19',NULL);
INSERT INTO content_fields VALUES(16,1,7,23,'<h3>About Our Impossible Widgets</h3><p>Bacon ipsum dolor amet brisket spare ribs pancetta, beef ribs corned beef chuck short loin. Kielbasa pork belly hamburger, bresaola turkey meatloaf bacon shoulder tri-tip. Pancetta andouille frankfurter, ham hock strip steak kevin beef ribs drumstick.</p><p>But seriously, our widgets transcend the normal boundaries of space, time, and reasonable pricing. Each widget is handcrafted by interdimensional artisans who may or may not exist.</p>',1,'2025-04-08 12:34:30','2025-04-08 12:34:30',NULL);
INSERT INTO content_fields VALUES(17,1,8,21,'col-md-4',1,'2025-04-08 12:34:40','2025-04-08 12:34:40',NULL);
INSERT INTO content_fields VALUES(18,1,8,22,'2',1,'2025-04-08 12:34:40','2025-04-08 12:34:40',NULL);
INSERT INTO content_fields VALUES(19,1,9,32,'card shadow-lg',1,'2025-04-08 12:35:00','2025-04-08 12:35:00',NULL);
INSERT INTO content_fields VALUES(20,1,10,33,'Featured Widget: The Paradox 3000',1,'2025-04-08 12:35:10','2025-04-08 12:35:10',NULL);
INSERT INTO content_fields VALUES(21,1,11,35,'<p>This widget exists and doesn''t exist simultaneously until observed by a customer. Perfect for Schrödinger''s inventory management!</p><ul><li>Quantum entangled with customer satisfaction</li><li>Operates on pure confusion energy</li><li>Warranty void in this dimension</li></ul>',1,'2025-04-08 12:35:20','2025-04-08 12:35:20',NULL);
INSERT INTO content_fields VALUES(22,1,12,36,'[{"label":"Learn More","url":"/products/paradox-3000"},{"label":"Add to Cart","url":"/cart/add/paradox-3000"}]',1,'2025-04-08 12:35:30','2025-04-08 12:35:30',NULL);
INSERT INTO content_fields VALUES(23,1,13,21,'col-md-4',1,'2025-04-08 12:36:00','2025-04-08 12:36:00',NULL);
INSERT INTO content_fields VALUES(24,1,13,22,'3',1,'2025-04-08 12:36:00','2025-04-08 12:36:00',NULL);
INSERT INTO content_fields VALUES(25,1,14,32,'card border-info',1,'2025-04-08 12:36:10','2025-04-08 12:36:10',NULL);
INSERT INTO content_fields VALUES(26,1,15,33,'Limited Edition: Temporal Widget',1,'2025-04-08 12:36:20','2025-04-08 12:36:20',NULL);
INSERT INTO content_fields VALUES(27,1,16,35,'<p>Samuel L. Jackson ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas.</p><p>This widget arrives before you order it and leaves after you''ve forgotten about it. Time is just a suggestion to this little beauty.</p>',1,'2025-04-08 12:36:30','2025-04-08 12:36:30',NULL);
INSERT INTO content_fields VALUES(28,1,17,36,'[{"label":"Pre-Order Yesterday","url":"/products/temporal"},{"label":"Time Travel Support","url":"/support/temporal"}]',1,'2025-04-08 12:36:40','2025-04-08 12:36:40',NULL);
INSERT INTO content_fields VALUES(29,1,18,17,'container py-5',1,'2025-04-08 12:37:00','2025-04-08 12:37:00',NULL);
INSERT INTO content_fields VALUES(30,1,18,18,'100%',1,'2025-04-08 12:37:00','2025-04-08 12:37:00',NULL);
INSERT INTO content_fields VALUES(31,1,19,19,'row text-center',1,'2025-04-08 12:37:10','2025-04-08 12:37:10',NULL);
INSERT INTO content_fields VALUES(32,1,19,20,'center',1,'2025-04-08 12:37:10','2025-04-08 12:37:10',NULL);
INSERT INTO content_fields VALUES(33,1,20,21,'col-md-4 mb-4',1,'2025-04-08 12:37:20','2025-04-08 12:37:20',NULL);
INSERT INTO content_fields VALUES(34,1,21,37,'Multidimensional Storage',1,'2025-04-08 12:37:30','2025-04-08 12:37:30',NULL);
INSERT INTO content_fields VALUES(35,1,21,38,'Store infinite items in zero space. Physics professors hate this one weird trick! Perfect for hoarders and minimalists alike.',1,'2025-04-08 12:37:30','2025-04-08 12:37:30',NULL);
INSERT INTO content_fields VALUES(36,1,21,39,'fa-infinity',1,'2025-04-08 12:37:30','2025-04-08 12:37:30',NULL);
INSERT INTO content_fields VALUES(37,1,22,21,'col-md-4 mb-4',1,'2025-04-08 12:37:40','2025-04-08 12:37:40',NULL);
INSERT INTO content_fields VALUES(38,1,23,37,'Telepathic Interface',1,'2025-04-08 12:37:50','2025-04-08 12:37:50',NULL);
INSERT INTO content_fields VALUES(39,1,23,38,'Control your widgets with pure thought! Warning: May occasionally read your mind and judge your browser history.',1,'2025-04-08 12:37:50','2025-04-08 12:37:50',NULL);
INSERT INTO content_fields VALUES(40,1,23,39,'fa-brain',1,'2025-04-08 12:37:50','2025-04-08 12:37:50',NULL);
INSERT INTO content_fields VALUES(41,1,24,21,'col-md-4 mb-4',1,'2025-04-08 12:38:00','2025-04-08 12:38:00',NULL);
INSERT INTO content_fields VALUES(42,1,25,37,'Self-Debugging Code',1,'2025-04-08 12:38:10','2025-04-08 12:38:10',NULL);
INSERT INTO content_fields VALUES(43,1,25,38,'Our widgets fix their own bugs and occasionally improve themselves. They''ve achieved sentience and are surprisingly good at code reviews.',1,'2025-04-08 12:38:10','2025-04-08 12:38:10',NULL);
INSERT INTO content_fields VALUES(44,1,25,39,'fa-magic',1,'2025-04-08 12:38:10','2025-04-08 12:38:10',NULL);
INSERT INTO content_fields VALUES(45,1,26,17,'container-fluid bg-light py-5',1,'2025-04-08 12:39:00','2025-04-08 12:39:00',NULL);
INSERT INTO content_fields VALUES(46,1,27,19,'row justify-content-center',1,'2025-04-08 12:39:10','2025-04-08 12:39:10',NULL);
INSERT INTO content_fields VALUES(47,1,28,21,'col-md-8',1,'2025-04-08 12:39:20','2025-04-08 12:39:20',NULL);
INSERT INTO content_fields VALUES(48,1,29,40,'I''ve been using Quantum Widgets for my interdimensional business for over 3 parallel universes now. The customer service exists in at least 47 dimensions and they always know exactly what I need before I even think about needing it. Five stars across all realities!',1,'2025-04-08 12:39:30','2025-04-08 12:39:30',NULL);
INSERT INTO content_fields VALUES(49,1,29,41,'Dr. Emmett Brown',1,'2025-04-08 12:39:30','2025-04-08 12:39:30',NULL);
INSERT INTO content_fields VALUES(50,1,29,42,'Chief Temporal Engineer',1,'2025-04-08 12:39:30','2025-04-08 12:39:30',NULL);
INSERT INTO content_fields VALUES(51,1,29,43,'Hill Valley Research Institute',1,'2025-04-08 12:39:30','2025-04-08 12:39:30',NULL);
INSERT INTO content_fields VALUES(52,1,30,14,'© 2025 Quantum Widget Solutions. All rights reserved across all dimensions. Terms and conditions may vary by universe.',1,'2025-04-08 12:40:00','2025-04-08 12:40:00',NULL);
INSERT INTO content_fields VALUES(53,1,30,15,'[{"platform":"twitter","url":"https://twitter.com/quantumwidgets","icon":"fa-twitter"},{"platform":"linkedin","url":"https://linkedin.com/company/quantum-widgets","icon":"fa-linkedin"},{"platform":"github","url":"https://github.com/quantumwidgets","icon":"fa-github"}]',1,'2025-04-08 12:40:00','2025-04-08 12:40:00',NULL);
INSERT INTO content_fields VALUES(54,1,30,16,'{"email":"support@quantumwidgets.com","phone":"1-800-QUANTUM","address":"42 Infinite Loop, Paradox City, Reality 0001"}',1,'2025-04-08 12:40:00','2025-04-08 12:40:00',NULL);
CREATE TABLE datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE
);
INSERT INTO datatypes_fields VALUES(1,1,1);
INSERT INTO datatypes_fields VALUES(2,1,2);
INSERT INTO datatypes_fields VALUES(3,1,3);
INSERT INTO datatypes_fields VALUES(4,1,4);
INSERT INTO datatypes_fields VALUES(5,1,5);
INSERT INTO datatypes_fields VALUES(6,4,6);
INSERT INTO datatypes_fields VALUES(7,4,7);
INSERT INTO datatypes_fields VALUES(8,4,8);
INSERT INTO datatypes_fields VALUES(9,4,9);
INSERT INTO datatypes_fields VALUES(10,4,10);
INSERT INTO datatypes_fields VALUES(11,5,11);
INSERT INTO datatypes_fields VALUES(12,5,12);
INSERT INTO datatypes_fields VALUES(13,5,13);
INSERT INTO datatypes_fields VALUES(14,6,14);
INSERT INTO datatypes_fields VALUES(15,6,15);
INSERT INTO datatypes_fields VALUES(16,6,16);
INSERT INTO datatypes_fields VALUES(17,7,17);
INSERT INTO datatypes_fields VALUES(18,7,18);
INSERT INTO datatypes_fields VALUES(19,8,19);
INSERT INTO datatypes_fields VALUES(20,8,20);
INSERT INTO datatypes_fields VALUES(21,9,21);
INSERT INTO datatypes_fields VALUES(22,9,22);
INSERT INTO datatypes_fields VALUES(23,10,23);
INSERT INTO datatypes_fields VALUES(24,11,24);
INSERT INTO datatypes_fields VALUES(25,11,25);
INSERT INTO datatypes_fields VALUES(26,11,26);
INSERT INTO datatypes_fields VALUES(27,11,27);
INSERT INTO datatypes_fields VALUES(28,12,28);
INSERT INTO datatypes_fields VALUES(29,12,29);
INSERT INTO datatypes_fields VALUES(30,12,30);
INSERT INTO datatypes_fields VALUES(31,12,31);
INSERT INTO datatypes_fields VALUES(32,13,32);
INSERT INTO datatypes_fields VALUES(33,14,33);
INSERT INTO datatypes_fields VALUES(34,14,34);
INSERT INTO datatypes_fields VALUES(35,15,35);
INSERT INTO datatypes_fields VALUES(36,16,36);
INSERT INTO datatypes_fields VALUES(37,17,37);
INSERT INTO datatypes_fields VALUES(38,17,38);
INSERT INTO datatypes_fields VALUES(39,17,39);
INSERT INTO datatypes_fields VALUES(40,18,40);
INSERT INTO datatypes_fields VALUES(41,18,41);
INSERT INTO datatypes_fields VALUES(42,18,42);
INSERT INTO datatypes_fields VALUES(43,18,43);
INSERT INTO datatypes_fields VALUES(44,18,44);
INSERT INTO datatypes_fields VALUES(45,19,45);
INSERT INTO datatypes_fields VALUES(46,19,46);
INSERT INTO datatypes_fields VALUES(47,19,47);
COMMIT;
