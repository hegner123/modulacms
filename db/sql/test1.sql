PRAGMA foreign_keys = ON;

INSERT INTO admin_datatype (admin_dt_id,adminrouteid,parentid,label,"type",author,authorid,datecreated,datemodified) VALUES
	 (1,1,NULL,'Parent','text','system',1,'LOCALTIMESTAMP','LOCALTIMESTAMP'),
	 (2,1,1,'Child','text','system',1,'LOCALTIMESTAMP','LOCALTIMESTAMP');
INSERT INTO admin_field (admin_field_id,adminrouteid,parentid,label,"data","type",author,authorid,datecreated,datemodified) VALUES
	 (1,1,1,'Title','ModulaCMS','text','system',1,'1209387123','128963419'),
	 (2,1,2,'Body','Climb leg meow meow you are my owner so here is a dead bird for under the bed sees bird in air, breaks into cage and attacks creature. Proudly present butt to human. Need to chase tail try to hold own back foot to clean it but foot reflexively kicks you in face, go into a rage and bite own foot, hard. Love to play with owner''s hair tie i like fish find box a little too small and curl up with fur hanging out mew mew stand in doorway, unwilling to chose whether to stay in or go out steal the warm chair right after you get up. ','richtext','system',1,'91823791823','19283917129');
INSERT INTO adminroute (admin_route_id,author,authorid,slug,title,status,datecreated,datemodified,template) VALUES
	 (1,'system',1,'/admin/','ModulaCMS',0,'1289476912','18274619','modula_base.html'),
	 (2,'system',1,'/admin/routes','ModulaCMS',0,'8924710927','17210947','modula_base.html'),
	 (3,'system',1,'/admin/datatypes','ModulaCMS',0,'6192873672','128746819','modula_base.html');
INSERT INTO datatype (datatype_id,routeid,parentid,label,"type",author,authorid,datecreated,datemodified) VALUES
	 (1,1,NULL,'PageBody','body','system',1,'91872312873','912873192'),
	 (2,1,1,'Hero','body','system',1,'91872312873','912873192'),
	 (3,1,1,'Content','body','system',1,'91872312873','912873192'),
	 (4,1,1,'Footer','body','system',1,'91872312873','912873192');
INSERT INTO field (field_id,routeid,parentid,label,"data","type",author,authorid,datecreated,datemodified) VALUES
	 (1,1,1,'Image','image','image','system',1,'198273912','1982361982'),
	 (2,1,1,'HeroText','Climb leg meow meow','text','system',1,'198273912','1982361982'),
	 (3,1,2,'Heading','Cats','text','system',1,'1729387123','192873918273'),
	 (4,1,2,'Body','Reward the chosen human with a slow blink sleep nap and slap kitten brother with paw pounce on unsuspecting person. Drink from the toilet climb into cupboard and lick the salt off rice cakes poop in a handbag look delicious and drink the soapy mopping up water then puke giant foamy fur-balls or poop in a handbag look delicious and drink the soapy mopping up water then puke giant foamy fur-balls flee in terror at cucumber discovered on floor cat is love, cat is life.','text','system',1,'1729387123','192873918273');
INSERT INTO media_dimension (id,label,width,height) VALUES
	 (1,'Tablet0',1920,1080),
	 (2,'Tablet1',1920,1080),
	 (3,'Tablet2',1920,1080),
	 (4,'Tablet3',1920,1080),
	 (5,'Tablet4',1920,1080);
INSERT INTO route (route_id,author,authorid,slug,title,status,datecreated,datemodified,content) VALUES
	 (1,'system',1,'/get/home','Test Site',0,'18723972','18274981273',''),
	 (2,'system',1,'/get/about','Test Site',0,'18723972','18274981273',''),
	 (3,'system',1,'/get/contact','Test Site',0,'18723972','18274981273',''),
	 (4,'system',1,'/get/sponsors','Test Site',0,'18723972','18274981273','');
INSERT INTO "user" (user_id,datecreated,datemodified,username,name,email,hash,"role") VALUES
	 (1,'98450298365','19827409124','system','system','system@modulacms.com','ioweyoiquyteksdbvl','admin');

INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
    'media1',
    'Media One',
    'Alt text for media one',
    'Caption for media one',
    'Description for media one',
    'class1',
    'system',
    1,
    '2023-10-01 12:00:00',
    '2023-10-01 12:00:00',
    'http://example.com/media1',
    'image/jpeg',
    '800x600',
    'media1_mobile.jpg',
    'media1_tablet.jpg',
    'media1_desktop.jpg',
    'media1_ultrawide.jpg'
);

INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
    'media2',
    'Media Two',
    'Alt text for media two',
    'Caption for media two',
    'Description for media two',
    'class2',
    'system',
    1,
    '2023-10-02 13:00:00',
    '2023-10-02 13:00:00',
    'http://example.com/media2',
    'image/png',
    '1024x768',
    'media2_mobile.png',
    'media2_tablet.png',
    'media2_desktop.png',
    'media2_ultrawide.png'
);

INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
    'media3',
    'Media Three',
    'Alt text for media three',
    'Caption for media three',
    'Description for media three',
    'class3',
    'system',
    1,
    '2023-10-03 14:00:00',
    '2023-10-03 14:00:00',
    'http://example.com/media3',
    'video/mp4',
    '1920x1080',
    'media3_mobile.mp4',
    'media3_tablet.mp4',
    'media3_desktop.mp4',
    'media3_ultrawide.mp4'
);

INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
    'media4',
    'Media Four',
    'Alt text for media four',
    'Caption for media four',
    'Description for media four',
    'class4',
    'system',
    1,
    '2023-10-04 15:00:00',
    '2023-10-04 15:00:00',
    'http://example.com/media4',
    'audio/mpeg',
    'N/A',
    'media4_mobile.mp3',
    'media4_tablet.mp3',
    'media4_desktop.mp3',
    'media4_ultrawide.mp3'
);

INSERT INTO media (
    name,
    displayname,
    alt,
    caption,
    description,
    class,
    author,
    authorid,
    datecreated,
    datemodified,
    url,
    mimetype,
    dimensions,
    optimizedmobile,
    optimizedtablet,
    optimizeddesktop,
    optimizedultrawide
) VALUES (
    'media5',
    'Media Five',
    'Alt text for media five',
    'Caption for media five',
    'Description for media five',
    'class5',
    'system',
    1,
    '2023-10-05 16:00:00',
    '2023-10-05 16:00:00',
    'http://example.com/media5',
    'application/pdf',
    'A4',
    'media5_mobile.pdf',
    'media5_tablet.pdf',
    'media5_desktop.pdf',
    'media5_ultrawide.pdf'
);

