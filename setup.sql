-- ModulaCMS Setup Script
-- This script creates all necessary tables and populates them with placeholder data

PRAGMA foreign_keys = ON;

-- ==================
-- Schema Definitions
-- ==================

-- Roles table for user permissions
CREATE TABLE IF NOT EXISTS roles (
    role_id INTEGER PRIMARY KEY,
    label TEXT NOT NULL UNIQUE,
    permissions TEXT NOT NULL UNIQUE
);

-- Users table for authentication and user management
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    hash TEXT NOT NULL,
    role INTEGER NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for user sessions
CREATE TABLE IF NOT EXISTS sessions (
    session_id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT,
    last_access TEXT DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    session_data TEXT
);

-- Authentication tokens
CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_type TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    revoked BOOLEAN DEFAULT 0
);

-- OAuth user connections
CREATE TABLE IF NOT EXISTS user_oauth (
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

-- Routes for content URLs
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Admin routes for admin panel URLs
CREATE TABLE IF NOT EXISTS admin_routes (
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

-- Tables definitions
CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    label TEXT UNIQUE,
    author_id INTEGER DEFAULT 1 NOT NULL
);

-- Content datatypes
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id INTEGER PRIMARY KEY,
    route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Admin datatypes
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- Field definitions
CREATE TABLE IF NOT EXISTS fields (
    field_id INTEGER PRIMARY KEY,
    route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Admin field definitions
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER DEFAULT NULL,
    parent_id INTEGER DEFAULT NULL,
    label TEXT NOT NULL,
    data TEXT,
    type TEXT NOT NULL,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- Content data entries
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL,
    datatype_id INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Content field values
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id INTEGER NOT NULL,
    content_data_id INTEGER NOT NULL,
    field_id INTEGER NOT NULL,
    field_value TEXT NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Admin content data entries
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER NOT NULL,
    admin_datatype_id INTEGER NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Admin content field values
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id INTEGER PRIMARY KEY,
    admin_route_id INTEGER NOT NULL,
    admin_content_data_id INTEGER NOT NULL,
    admin_field_id INTEGER NOT NULL,
    admin_field_value TEXT NOT NULL,
    history TEXT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- Media dimensions
CREATE TABLE IF NOT EXISTS media_dimensions (
    md_id INTEGER PRIMARY KEY,
    label TEXT,
    width INTEGER,
    height INTEGER,
    aspect_ratio TEXT
);

-- Media assets
CREATE TABLE IF NOT EXISTS media (
    media_id INTEGER PRIMARY KEY,
    name TEXT,
    display_name TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    author TEXT DEFAULT 'system' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    mimetype TEXT,
    dimensions TEXT,
    url TEXT,
    optimized_mobile TEXT,
    optimized_tablet TEXT,
    optimized_desktop TEXT,
    optimized_ultra_wide TEXT
);

-- ==================
-- Initial Data
-- ==================

-- Default roles must be inserted first because users depend on them
INSERT INTO roles (role_id, label, permissions) VALUES 
(1, 'admin', '{"admin":true,"create":true,"read":true,"update":true,"delete":true}');

INSERT INTO roles (role_id, label, permissions) VALUES 
(2, 'editor', '{"admin":false,"create":true,"read":true,"update":true,"delete":false}');

INSERT INTO roles (role_id, label, permissions) VALUES 
(3, 'author', '{"admin":false,"create":true,"read":true,"update":true,"delete":false}');

INSERT INTO roles (role_id, label, permissions) VALUES 
(4, 'contributor', '{"admin":false,"create":true,"read":true,"update":false,"delete":false}');

INSERT INTO roles (role_id, label, permissions) VALUES 
(5, 'subscriber', '{"admin":false,"create":false,"read":true,"update":false,"delete":false}');

-- Default admin user with bcrypt hashed password 'admin123'
-- Password is hashed using bcrypt with cost=12
INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES 
(1, 'admin', 'System Administrator', 'admin@example.com', '$2a$12$kZJ9.UG8cZ5oM4XB5IFk0uHGFCrL7dR.Cg9VeJDYQ/HU/V07yvVLq', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Demo user with bcrypt hashed password 'demo123'
INSERT INTO users (user_id, username, name, email, hash, role, date_created, date_modified) VALUES 
(2, 'demo', 'Demo User', 'demo@example.com', '$2a$12$ZxJZLrfE7UGVlnHPrJsOR.k94Z0QG6TFnGXYSdwVcRtE8wE2NbL2y', 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default tables
INSERT INTO tables (id, label, author_id) VALUES 
(1, 'users', 1),
(2, 'roles', 1),
(3, 'routes', 1),
(4, 'content', 1);

-- Media dimensions
INSERT INTO media_dimensions (md_id, label, width, height, aspect_ratio) VALUES 
(1, 'Mobile', 480, 320, '3:2'),
(2, 'Tablet', 800, 600, '4:3'),
(3, 'Desktop', 1920, 1080, '16:9'),
(4, 'UltraWide', 3440, 1440, '21:9');

-- Default routes
INSERT INTO routes (route_id, author, author_id, slug, title, status, date_created, date_modified) VALUES 
(1, 'admin', 1, 'home', 'Home', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 'admin', 1, 'about', 'About Us', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, 'admin', 1, 'contact', 'Contact Us', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 'admin', 1, 'blog', 'Blog', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default admin routes
INSERT INTO admin_routes (admin_route_id, slug, title, status, author, author_id, date_created, date_modified) VALUES 
(1, 'dashboard', 'Dashboard', 1, 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 'users', 'User Management', 1, 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, 'content', 'Content Management', 1, 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 'settings', 'System Settings', 1, 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default datatypes
INSERT INTO datatypes (datatype_id, route_id, parent_id, label, type, author, author_id, date_created, date_modified) VALUES 
(1, 1, NULL, 'Page', 'page', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 4, NULL, 'Post', 'post', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, NULL, NULL, 'Media', 'media', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, NULL, NULL, 'Widget', 'widget', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default admin datatypes
INSERT INTO admin_datatypes (admin_datatype_id, admin_route_id, parent_id, label, type, author, author_id, date_created, date_modified) VALUES 
(1, 3, NULL, 'Page Editor', 'page-editor', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 3, NULL, 'Post Editor', 'post-editor', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, 2, NULL, 'User Editor', 'user-editor', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 4, NULL, 'Settings Editor', 'settings-editor', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default fields
INSERT INTO fields (field_id, route_id, parent_id, label, data, type, author, author_id, date_created, date_modified) VALUES 
(1, NULL, NULL, 'Title', '{"required":true,"default":""}', 'text', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, NULL, NULL, 'Content', '{"required":true,"default":""}', 'richtext', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, NULL, NULL, 'Featured Image', '{"required":false,"default":""}', 'image', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, NULL, NULL, 'Tags', '{"required":false,"default":"","multiple":true}', 'tag', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(5, NULL, NULL, 'Publish Date', '{"required":true,"default":"now"}', 'datetime', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Default admin fields
INSERT INTO admin_fields (admin_field_id, admin_route_id, parent_id, label, data, type, author, author_id, date_created, date_modified) VALUES 
(1, 3, NULL, 'Username', '{"required":true,"minLength":3,"maxLength":50}', 'text', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 3, NULL, 'Email', '{"required":true,"format":"email"}', 'email', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(3, 3, NULL, 'Password', '{"required":true,"minLength":8}', 'password', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 3, NULL, 'Role', '{"required":true,"options":"roles"}', 'select', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(5, 4, NULL, 'Site Title', '{"required":true}', 'text', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content data for Home page
INSERT INTO content_data (content_data_id, route_id, datatype_id, date_created, date_modified) VALUES 
(1, 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content fields for Home page
INSERT INTO content_fields (content_field_id, route_id, content_data_id, field_id, field_value, date_created, date_modified) VALUES 
(1, 1, 1, 1, 'Welcome to ModulaCMS', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(2, 1, 1, 2, '<h1>Welcome to ModulaCMS</h1><p>This is a flexible content management system that allows you to create and manage your website content easily.</p><p>Start by navigating to the admin panel to create your first post or page.</p>', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content data for About page
INSERT INTO content_data (content_data_id, route_id, datatype_id, date_created, date_modified) VALUES 
(2, 2, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content fields for About page
INSERT INTO content_fields (content_field_id, route_id, content_data_id, field_id, field_value, date_created, date_modified) VALUES 
(3, 2, 2, 1, 'About Our Company', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(4, 2, 2, 2, '<h1>About ModulaCMS</h1><p>ModulaCMS is a modern content management system built with Go and designed for flexibility and performance.</p><p>Our team is dedicated to creating tools that make web publishing accessible to everyone.</p>', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content data for Blog post
INSERT INTO content_data (content_data_id, route_id, datatype_id, date_created, date_modified) VALUES 
(3, 4, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample content fields for Blog post
INSERT INTO content_fields (content_field_id, route_id, content_data_id, field_id, field_value, date_created, date_modified) VALUES 
(5, 4, 3, 1, 'Getting Started with ModulaCMS', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(6, 4, 3, 2, '<h1>Getting Started with ModulaCMS</h1><p>In this post, we will explore the basics of setting up and using ModulaCMS for your website or application.</p><p>ModulaCMS provides a powerful and flexible foundation for creating dynamic web content with minimal effort.</p>', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(7, 4, 3, 4, 'tutorial,beginners,setup', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(8, 4, 3, 5, '2023-01-15 10:30:00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample admin content data for dashboard
INSERT INTO admin_content_data (admin_content_data_id, admin_route_id, admin_datatype_id, date_created, date_modified) VALUES 
(1, 1, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample admin content fields for dashboard
INSERT INTO admin_content_fields (admin_content_field_id, admin_route_id, admin_content_data_id, admin_field_id, admin_field_value, date_created, date_modified) VALUES 
(1, 1, 1, 5, 'ModulaCMS Dashboard', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Sample media entry
INSERT INTO media (media_id, name, display_name, alt, caption, description, mimetype, url, author, author_id, date_created, date_modified) VALUES 
(1, 'sample-image.jpg', 'Sample Image', 'A sample image', 'Sample image caption', 'This is a sample image for demonstration purposes', 'image/jpeg', '/media/sample-image.jpg', 'admin', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);