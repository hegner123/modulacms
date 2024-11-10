package main

const mediaTableConst string = `CREATE TABLE IF NOT EXISTS media (id INTEGER PRIMARY KEY, name TEXT NOT NULL, displayname TEXT, alt TEXT, caption TEXT, description TEXT, class TEXT, author TEXT, authorid INTEGER, datecreated TEXT, datemodified TEXT, url TEXT, mimeType TEXT, dimensions TEXT, optimizedmobile TEXT, optimizedtablet TEXT, optimizeddesktop TEXT, optimizedultrawide TEXT);`
const userTableConst string = `CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, datecreated TEXT, datemodified TEXT, username TEXT, name TEXT, email TEXT, hash TEXT, role TEXT);`
const adminRoutesTableConst string = `CREATE TABLE IF NOT EXISTS adminroutes (id INTEGER PRIMARY KEY, slug TEXT NOT NULL, author TEXT, authorId INTEGER, title TEXT, status INTEGER NOT NULL, datecreated TEXT NOT NULL, datemodified TEXT NOT NULL, content TEXT NOT NULL, type TEXT NOT NULL, template TEXT);`
const routesTableConst string = `CREATE TABLE IF NOT EXISTS routes (id INTEGER PRIMARY KEY, slug TEXT NOT NULL, author TEXT, authorId INTEGER, title TEXT, status INTEGER NOT NULL, datecreated TEXT NOT NULL, datemodified TEXT NOT NULL, content TEXT NOT NULL, type TEXT NOT NULL, template TEXT);`
const fieldsTableConst string = `CREATE TABLE IF NOT EXISTS fields(id INTEGER PRIMARY KEY, routeId INTEGER NOT NULL, author TEXT, authorId TEXT, key TEXT, type TEXT, data TEXT, datecreated TEXT, datemodified TEXT, component TEXT, tags TEXT, parent TEXT);`
const tables string = "CREATE TABLE IF NOT EXISTS tables (id INTEGER PRIMARY KEY, label TEXT UNIQUE);"
var menuElement = Element{Tag: "menu-component",Attributes: map[string]string{"id":"0"}}
var createFieldField Field = Field{RouteID: 0, Author: "system", AuthorID: "0", Key: "Menu", Type: "Menu", Data: "", DateCreated: Times, DateModified: Times, Component:menuElement, Tags: "Menu, menu, Navigation, navigation, Nav, nav", Parent: "root" }
