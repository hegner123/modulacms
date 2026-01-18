-- ModulaCMS Demo Data
-- This file contains sample data for all tables demonstrating the tree structure and relationships

-- =============================================================================
-- 1. PERMISSIONS
-- =============================================================================
INSERT INTO permissions (permission_id, table_id, mode, label) VALUES
(1, 1, 1, 'View Users'),
(2, 1, 2, 'Create Users'),
(3, 1, 4, 'Update Users'),
(4, 1, 8, 'Delete Users'),
(5, 2, 1, 'View Content'),
(6, 2, 2, 'Create Content'),
(7, 2, 4, 'Update Content'),
(8, 2, 8, 'Delete Content'),
(9, 3, 1, 'View Media'),
(10, 3, 2, 'Upload Media'),
(11, 3, 4, 'Update Media'),
(12, 3, 8, 'Delete Media');

-- =============================================================================
-- 2. ROLES
-- =============================================================================
INSERT INTO roles (role_id, label, permissions) VALUES
(1, 'Super Admin', '1,2,3,4,5,6,7,8,9,10,11,12'),
(2, 'Editor', '5,6,7,9,10,11'),
(3, 'Contributor', '5,6,9,10'),
(4, 'Viewer', '1,5,9');

-- =============================================================================
-- 3. MEDIA_DIMENSIONS
-- =============================================================================
INSERT INTO media_dimensions (md_id, label, width, height, aspect_ratio) VALUES
(1, 'Mobile', 480, 320, '3:2'),
(2, 'Tablet', 1024, 768, '4:3'),
(3, 'Desktop', 1920, 1080, '16:9'),
(4, 'Ultra Wide', 3440, 1440, '21:9'),
(5, 'Thumbnail', 200, 200, '1:1'),
(6, 'Hero', 2560, 1440, '16:9');

-- =============================================================================
-- 4. USERS
-- =============================================================================
INSERT INTO users (user_id, username, name, email, hash, role, date_created) VALUES
(1, 'system', 'System User', 'system@modulacms.local', '$2a$10$systemhashexample', 1, datetime('now')),
(2, 'admin', 'Admin User', 'admin@example.com', '$2a$10$adminhashexample', 1, datetime('now')),
(3, 'editor', 'Jane Editor', 'jane@example.com', '$2a$10$editorhashexample', 2, datetime('now')),
(4, 'contributor', 'Bob Writer', 'bob@example.com', '$2a$10$contributorhashexample', 3, datetime('now')),
(5, 'viewer', 'Alice Reader', 'alice@example.com', '$2a$10$viewerhashexample', 4, datetime('now'));

-- =============================================================================
-- 5. ADMIN_ROUTES
-- =============================================================================
INSERT INTO admin_routes (admin_route_id, slug, title, status, author_id, date_created) VALUES
(1, '/admin/dashboard', 'Dashboard', 1, 1, datetime('now')),
(2, '/admin/users', 'User Management', 1, 1, datetime('now')),
(3, '/admin/content', 'Content Management', 1, 1, datetime('now')),
(4, '/admin/media', 'Media Library', 1, 1, datetime('now')),
(5, '/admin/settings', 'Settings', 1, 1, datetime('now'));

-- =============================================================================
-- 6. ROUTES
-- =============================================================================
INSERT INTO routes (route_id, slug, title, status, author_id, date_created) VALUES
(1, '/', 'Home', 1, 2, datetime('now')),
(2, '/about', 'About Us', 1, 2, datetime('now')),
(3, '/services', 'Services', 1, 2, datetime('now')),
(4, '/services/web-development', 'Web Development', 1, 3, datetime('now')),
(5, '/services/consulting', 'Consulting', 1, 3, datetime('now')),
(6, '/blog', 'Blog', 1, 3, datetime('now')),
(7, '/contact', 'Contact', 1, 2, datetime('now'));

-- =============================================================================
-- 7. DATATYPES
-- =============================================================================
INSERT INTO datatypes (datatype_id, parent_id, label, type, author_id, date_created) VALUES
(1, NULL, 'Page', 'page', 1, datetime('now')),
(2, NULL, 'Article', 'article', 1, datetime('now')),
(3, NULL, 'Service', 'service', 1, datetime('now')),
(4, 2, 'Blog Post', 'blog_post', 1, datetime('now')),
(5, NULL, 'Contact Form', 'form', 1, datetime('now'));

-- =============================================================================
-- 8. FIELDS
-- =============================================================================
INSERT INTO fields (field_id, parent_id, label, data, type, author_id, date_created) VALUES
(1, 1, 'Page Title', '', 'text', 1, datetime('now')),
(2, 1, 'Page Content', '', 'richtext', 1, datetime('now')),
(3, 1, 'Meta Description', '', 'textarea', 1, datetime('now')),
(4, 2, 'Article Title', '', 'text', 1, datetime('now')),
(5, 2, 'Article Body', '', 'richtext', 1, datetime('now')),
(6, 2, 'Author Name', '', 'text', 1, datetime('now')),
(7, 2, 'Published Date', '', 'date', 1, datetime('now')),
(8, 3, 'Service Name', '', 'text', 1, datetime('now')),
(9, 3, 'Service Description', '', 'richtext', 1, datetime('now')),
(10, 3, 'Price', '', 'number', 1, datetime('now')),
(11, 5, 'Contact Email', '', 'email', 1, datetime('now')),
(12, 5, 'Message', '', 'textarea', 1, datetime('now'));

-- =============================================================================
-- 9. ADMIN_DATATYPES
-- =============================================================================
INSERT INTO admin_datatypes (admin_datatype_id, parent_id, label, type, author_id, date_created) VALUES
(1, NULL, 'Dashboard Widget', 'widget', 1, datetime('now')),
(2, NULL, 'User Profile', 'profile', 1, datetime('now')),
(3, NULL, 'Content Block', 'block', 1, datetime('now')),
(4, NULL, 'Media Item', 'media', 1, datetime('now'));

-- =============================================================================
-- 10. ADMIN_FIELDS
-- =============================================================================
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created) VALUES
(1, 1, 'Widget Title', '', 'text', 1, datetime('now')),
(2, 1, 'Widget Type', 'stats', 'select', 1, datetime('now')),
(3, 2, 'Display Name', '', 'text', 1, datetime('now')),
(4, 2, 'Bio', '', 'textarea', 1, datetime('now')),
(5, 3, 'Block Title', '', 'text', 1, datetime('now')),
(6, 3, 'Block Content', '', 'richtext', 1, datetime('now')),
(7, 4, 'File Name', '', 'text', 1, datetime('now')),
(8, 4, 'Alt Text', '', 'text', 1, datetime('now'));

-- =============================================================================
-- 11. TOKENS
-- =============================================================================
INSERT INTO tokens (id, user_id, token_type, token, issued_at, expires_at, revoked) VALUES
(1, 2, 'access', 'tok_admin_access_example123', datetime('now'), datetime('now', '+1 day'), 0),
(2, 3, 'access', 'tok_editor_access_example456', datetime('now'), datetime('now', '+1 day'), 0),
(3, 2, 'refresh', 'tok_admin_refresh_example789', datetime('now'), datetime('now', '+30 days'), 0);

-- =============================================================================
-- 12. USER_OAUTH
-- =============================================================================
INSERT INTO user_oauth (user_oauth_id, user_id, oauth_provider, oauth_provider_user_id, access_token, refresh_token, token_expires_at, date_created) VALUES
(1, 2, 'google', 'google_user_12345', 'ya29.example_access_token', '1//example_refresh_token_google', datetime('now', '+1 hour'), datetime('now')),
(2, 3, 'github', 'github_user_67890', 'gho_example_github_token', 'ghr_example_refresh_token_github', datetime('now', '+1 hour'), datetime('now'));

-- =============================================================================
-- 13. TABLES
-- =============================================================================
INSERT INTO tables (id, label, author_id) VALUES
(1, 'users', 1),
(2, 'content', 1),
(3, 'media', 1),
(4, 'routes', 1),
(5, 'datatypes', 1);

-- =============================================================================
-- 14. MEDIA
-- =============================================================================
INSERT INTO media (media_id, name, display_name, alt, caption, description, class, author_id, mimetype, dimensions, url, date_created) VALUES
(1, 'hero-banner.jpg', 'Hero Banner', 'Company hero banner image', 'Welcome to our site', 'Main hero banner for homepage', 'hero-image', 2, 'image/jpeg', '1920x1080', '/media/hero-banner.jpg', datetime('now')),
(2, 'about-team.jpg', 'Team Photo', 'Our amazing team', 'The team in 2025', 'Group photo of our team', 'team-photo', 2, 'image/jpeg', '1600x900', '/media/about-team.jpg', datetime('now')),
(3, 'service-web.jpg', 'Web Development', 'Web development services illustration', 'Custom web solutions', 'Web development service image', 'service-image', 3, 'image/jpeg', '800x600', '/media/service-web.jpg', datetime('now')),
(4, 'service-consulting.jpg', 'Consulting', 'Consulting services illustration', 'Expert consulting', 'Consulting service image', 'service-image', 3, 'image/jpeg', '800x600', '/media/service-consulting.jpg', datetime('now')),
(5, 'logo.svg', 'Company Logo', 'ModulaCMS logo', 'Our brand', 'Company logo in SVG format', 'logo', 1, 'image/svg+xml', '200x200', '/media/logo.svg', datetime('now'));

-- =============================================================================
-- 15. SESSIONS
-- =============================================================================
INSERT INTO sessions (session_id, user_id, created_at, expires_at, last_access, ip_address, user_agent, session_data) VALUES
(1, 2, datetime('now'), datetime('now', '+1 day'), datetime('now'), '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)', '{"preferences": {"theme": "dark"}}'),
(2, 3, datetime('now'), datetime('now', '+1 day'), datetime('now'), '192.168.1.101', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)', '{"preferences": {"theme": "light"}}');

-- =============================================================================
-- 16. CONTENT_DATA (with Tree Structure)
-- =============================================================================
-- Demonstrating tree structure:
-- Home (1)
--   ├── About Us (2)
--   ├── Services (3)
--   │   ├── Web Development (5)
--   │   └── Consulting (6)
--   ├── Blog (7)
--   │   ├── First Post (8)
--   │   └── Second Post (9)
--   └── Contact (4)

INSERT INTO content_data (content_data_id, parent_id, first_child_id, next_sibling_id, prev_sibling_id, route_id, datatype_id, author_id, date_created) VALUES
-- Root: Home (has children, no parent)
(1, NULL, 2, NULL, NULL, 1, 1, 2, datetime('now')),
-- Level 1: About Us (parent=1, next sibling=3)
(2, 1, NULL, 3, NULL, 2, 1, 2, datetime('now')),
-- Level 1: Services (parent=1, has children, prev=2, next=7)
(3, 1, 5, 7, 2, 3, 1, 2, datetime('now')),
-- Level 2: Web Development (parent=3, next sibling=6)
(5, 3, NULL, 6, NULL, 4, 3, 3, datetime('now')),
-- Level 2: Consulting (parent=3, prev sibling=5)
(6, 3, NULL, NULL, 5, 5, 3, 3, datetime('now')),
-- Level 1: Blog (parent=1, has children, prev=3, next=4)
(7, 1, 8, 4, 3, 6, 1, 3, datetime('now')),
-- Level 2: First Post (parent=7, next sibling=9)
(8, 7, NULL, 9, NULL, 6, 4, 3, datetime('now')),
-- Level 2: Second Post (parent=7, prev sibling=8)
(9, 7, NULL, NULL, 8, 6, 4, 3, datetime('now')),
-- Level 1: Contact (parent=1, prev sibling=7)
(4, 1, NULL, NULL, 7, 7, 5, 2, datetime('now'));

-- =============================================================================
-- 17. CONTENT_FIELDS
-- =============================================================================
INSERT INTO content_fields (content_field_id, route_id, content_data_id, field_id, field_value, author_id, date_created) VALUES
-- Home page fields
(1, 1, 1, 1, 'Welcome to ModulaCMS', 2, datetime('now')),
(2, 1, 1, 2, '<h1>Build Better Websites</h1><p>ModulaCMS is a modern headless CMS built with Go.</p>', 2, datetime('now')),
(3, 1, 1, 3, 'ModulaCMS - Modern headless CMS for agencies', 2, datetime('now')),
-- About Us fields
(4, 2, 2, 1, 'About Our Company', 2, datetime('now')),
(5, 2, 2, 2, '<p>We are a team of passionate developers building the next generation of content management.</p>', 2, datetime('now')),
(6, 2, 2, 3, 'Learn about our company and mission', 2, datetime('now')),
-- Services parent fields
(7, 3, 3, 1, 'Our Services', 2, datetime('now')),
(8, 3, 3, 2, '<p>We offer comprehensive web development and consulting services.</p>', 2, datetime('now')),
-- Web Development service
(9, 4, 5, 8, 'Web Development', 3, datetime('now')),
(10, 4, 5, 9, '<p>Custom web applications built with modern technologies.</p>', 3, datetime('now')),
(11, 4, 5, 10, '5000', 3, datetime('now')),
-- Consulting service
(12, 5, 6, 8, 'Technical Consulting', 3, datetime('now')),
(13, 5, 6, 9, '<p>Expert guidance on architecture and technology decisions.</p>', 3, datetime('now')),
(14, 5, 6, 10, '3000', 3, datetime('now')),
-- Blog parent
(15, 6, 7, 1, 'Blog', 3, datetime('now')),
(16, 6, 7, 2, '<p>Latest news and insights from our team.</p>', 3, datetime('now')),
-- Blog posts
(17, 6, 8, 4, 'Getting Started with ModulaCMS', 3, datetime('now')),
(18, 6, 8, 5, '<p>Learn how to build your first site with ModulaCMS...</p>', 3, datetime('now')),
(19, 6, 8, 6, 'Jane Editor', 3, datetime('now')),
(20, 6, 8, 7, '2025-01-10', 3, datetime('now')),
(21, 6, 9, 4, 'Advanced Tree Structures', 3, datetime('now')),
(22, 6, 9, 5, '<p>Deep dive into sibling-pointer trees...</p>', 3, datetime('now')),
(23, 6, 9, 6, 'Jane Editor', 3, datetime('now')),
(24, 6, 9, 7, '2025-01-12', 3, datetime('now')),
-- Contact
(25, 7, 4, 1, 'Get In Touch', 2, datetime('now')),
(26, 7, 4, 2, '<p>Contact us for more information.</p>', 2, datetime('now'));

-- =============================================================================
-- 18. ADMIN_CONTENT_DATA (with Tree Structure)
-- =============================================================================
-- Admin dashboard structure:
-- Dashboard (1)
--   ├── Stats Widget (2)
--   └── Activity Widget (3)
-- User Management (4)
--   ├── User List (5)
--   └── Role Management (6)

INSERT INTO admin_content_data (admin_content_data_id, parent_id, first_child_id, next_sibling_id, prev_sibling_id, admin_route_id, admin_datatype_id, author_id, date_created) VALUES
-- Root: Dashboard (has children)
(1, NULL, 2, NULL, NULL, 1, 1, 1, datetime('now')),
-- Level 1: Stats Widget (parent=1, next=3)
(2, 1, NULL, 3, NULL, 1, 1, 1, datetime('now')),
-- Level 1: Activity Widget (parent=1, prev=2)
(3, 1, NULL, NULL, 2, 1, 1, 1, datetime('now')),
-- Root: User Management (has children)
(4, NULL, 5, NULL, NULL, 2, 2, 1, datetime('now')),
-- Level 1: User List (parent=4, next=6)
(5, 4, NULL, 6, NULL, 2, 2, 1, datetime('now')),
-- Level 1: Role Management (parent=4, prev=5)
(6, 4, NULL, NULL, 5, 2, 2, 1, datetime('now'));

-- =============================================================================
-- 19. ADMIN_CONTENT_FIELDS
-- =============================================================================
INSERT INTO admin_content_fields (admin_content_field_id, admin_route_id, admin_content_data_id, admin_field_id, admin_field_value, author_id, date_created) VALUES
-- Dashboard widgets
(1, 1, 2, 1, 'User Statistics', 1, datetime('now')),
(2, 1, 2, 2, 'stats', 1, datetime('now')),
(3, 1, 3, 1, 'Recent Activity', 1, datetime('now')),
(4, 1, 3, 2, 'activity', 1, datetime('now')),
-- User management
(5, 2, 5, 3, 'All Users', 1, datetime('now')),
(6, 2, 5, 4, 'Comprehensive list of all system users', 1, datetime('now')),
(7, 2, 6, 3, 'Role Management', 1, datetime('now')),
(8, 2, 6, 4, 'Manage user roles and permissions', 1, datetime('now'));

-- =============================================================================
-- 20. DATATYPES_FIELDS (Junction Table)
-- =============================================================================
INSERT INTO datatypes_fields (id, datatype_id, field_id) VALUES
-- Page datatype uses title, content, meta fields
(1, 1, 1),
(2, 1, 2),
(3, 1, 3),
-- Article datatype uses article-specific fields
(4, 2, 4),
(5, 2, 5),
(6, 2, 6),
(7, 2, 7),
-- Service datatype uses service fields
(8, 3, 8),
(9, 3, 9),
(10, 3, 10),
-- Blog Post inherits from Article
(11, 4, 4),
(12, 4, 5),
(13, 4, 6),
(14, 4, 7),
-- Contact Form uses contact fields
(15, 5, 11),
(16, 5, 12);

-- =============================================================================
-- 21. ADMIN_DATATYPES_FIELDS (Junction Table)
-- =============================================================================
INSERT INTO admin_datatypes_fields (id, admin_datatype_id, admin_field_id) VALUES
-- Dashboard Widget uses widget fields
(1, 1, 1),
(2, 1, 2),
-- User Profile uses profile fields
(3, 2, 3),
(4, 2, 4),
-- Content Block uses block fields
(5, 3, 5),
(6, 3, 6),
-- Media Item uses media fields
(7, 4, 7),
(8, 4, 8);

-- =============================================================================
-- DEMO DATA SUMMARY
-- =============================================================================
-- Total records inserted:
-- - Permissions: 12
-- - Roles: 4
-- - Media Dimensions: 6
-- - Users: 5
-- - Admin Routes: 5
-- - Routes: 7
-- - Datatypes: 5
-- - Fields: 12
-- - Admin Datatypes: 4
-- - Admin Fields: 8
-- - Tokens: 3
-- - User OAuth: 2
-- - Tables: 5
-- - Media: 5
-- - Sessions: 2
-- - Content Data: 9 (demonstrating tree structure)
-- - Content Fields: 26
-- - Admin Content Data: 6 (demonstrating tree structure)
-- - Admin Content Fields: 8
-- - Datatypes Fields: 16
-- - Admin Datatypes Fields: 8
--
-- Tree Structure Examples:
-- 1. Frontend Content Tree (content_data):
--    - 3-level hierarchy with sibling pointers
--    - Demonstrates parent-child and sibling relationships
--    - Root node (Home) with multiple children
--    - Branch node (Services) with sub-services
--
-- 2. Admin Content Tree (admin_content_data):
--    - Dashboard widgets organization
--    - User management hierarchy
--    - Demonstrates administrative content structure
