-- ModulaCMS Demo Data
-- This file contains comprehensive sample data for all tables
-- Order follows sql/create_order.md for proper foreign key dependencies

-- NOTE: permissions table is not in create_order.md but exists in schema
-- Including it first since roles references it
-- =============================================================================
-- 0. PERMISSIONS
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
(12, 3, 8, 'Delete Media'),
(13, 4, 1, 'View Routes'),
(14, 4, 2, 'Create Routes'),
(15, 4, 4, 'Update Routes'),
(16, 4, 8, 'Delete Routes'),
(17, 5, 1, 'View Datatypes'),
(18, 5, 2, 'Create Datatypes'),
(19, 5, 4, 'Update Datatypes'),
(20, 5, 8, 'Delete Datatypes');

-- =============================================================================
-- 1. ROLES
-- =============================================================================
INSERT INTO roles (role_id, label, permissions) VALUES
(1, 'Super Admin', '1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20'),
(2, 'Editor', '5,6,7,9,10,11,13,14,15,17,18,19'),
(3, 'Contributor', '5,6,9,10,13,17'),
(4, 'Viewer', '1,5,9,13,17'),
(5, 'Content Manager', '5,6,7,8,13,14,15'),
(6, 'Media Manager', '9,10,11,12'),
(7, 'Developer', '1,5,9,13,17,18,19,20');

-- =============================================================================
-- 2. MEDIA_DIMENSIONS
-- =============================================================================
INSERT INTO media_dimensions (md_id, label, width, height, aspect_ratio) VALUES
(1, 'Mobile', 480, 320, '3:2'),
(2, 'Tablet', 1024, 768, '4:3'),
(3, 'Desktop', 1920, 1080, '16:9'),
(4, 'Ultra Wide', 3440, 1440, '21:9'),
(5, 'Thumbnail', 200, 200, '1:1'),
(6, 'Hero', 2560, 1440, '16:9'),
(7, 'Portrait', 800, 1200, '2:3'),
(8, 'Square Large', 1200, 1200, '1:1'),
(9, 'Banner', 1920, 400, '24:5'),
(10, '4K', 3840, 2160, '16:9');

-- =============================================================================
-- 3. USERS
-- =============================================================================
INSERT INTO users (user_id, username, name, email, hash, role, date_created) VALUES
(1, 'system', 'System User', 'system@modulacms.local', '$2a$10$systemhashexample', 1, datetime('now', '-365 days')),
(2, 'admin', 'Admin User', 'admin@example.com', '$2a$10$adminhashexample', 1, datetime('now', '-180 days')),
(3, 'editor', 'Jane Editor', 'jane@example.com', '$2a$10$editorhashexample', 2, datetime('now', '-120 days')),
(4, 'contributor', 'Bob Writer', 'bob@example.com', '$2a$10$contributorhashexample', 3, datetime('now', '-90 days')),
(5, 'viewer', 'Alice Reader', 'alice@example.com', '$2a$10$viewerhashexample', 4, datetime('now', '-60 days')),
(6, 'jsmith', 'John Smith', 'john.smith@example.com', '$2a$10$jsmithhashexample', 2, datetime('now', '-45 days')),
(7, 'mgarcia', 'Maria Garcia', 'maria.garcia@example.com', '$2a$10$mgarciahashexample', 5, datetime('now', '-30 days')),
(8, 'dwilson', 'David Wilson', 'david.wilson@example.com', '$2a$10$dwilsonhashexample', 3, datetime('now', '-20 days')),
(9, 'sjohnson', 'Sarah Johnson', 'sarah.j@example.com', '$2a$10$sjohnsonhashexample', 6, datetime('now', '-15 days')),
(10, 'rbrown', 'Robert Brown', 'rbrown@example.com', '$2a$10$rbrownhashexample', 7, datetime('now', '-10 days'));

-- =============================================================================
-- 4. ADMIN_ROUTES
-- =============================================================================
INSERT INTO admin_routes (admin_route_id, slug, title, status, author_id, date_created) VALUES
(1, '/admin/dashboard', 'Dashboard', 1, 1, datetime('now', '-180 days')),
(2, '/admin/users', 'User Management', 1, 1, datetime('now', '-180 days')),
(3, '/admin/content', 'Content Management', 1, 1, datetime('now', '-180 days')),
(4, '/admin/media', 'Media Library', 1, 1, datetime('now', '-180 days')),
(5, '/admin/settings', 'Settings', 1, 1, datetime('now', '-180 days')),
(6, '/admin/analytics', 'Analytics', 1, 2, datetime('now', '-90 days')),
(7, '/admin/plugins', 'Plugin Manager', 1, 2, datetime('now', '-60 days')),
(8, '/admin/themes', 'Theme Settings', 1, 2, datetime('now', '-60 days')),
(9, '/admin/backup', 'Backup & Restore', 1, 1, datetime('now', '-30 days')),
(10, '/admin/logs', 'System Logs', 1, 10, datetime('now', '-15 days'));

-- =============================================================================
-- 5. ROUTES
-- =============================================================================
INSERT INTO routes (route_id, slug, title, status, author_id, date_created) VALUES
(1, '/', 'Home', 1, 2, datetime('now', '-120 days')),
(2, '/about', 'About Us', 1, 2, datetime('now', '-120 days')),
(3, '/services', 'Services', 1, 2, datetime('now', '-120 days')),
(4, '/services/web-development', 'Web Development', 1, 3, datetime('now', '-110 days')),
(5, '/services/consulting', 'Consulting', 1, 3, datetime('now', '-110 days')),
(6, '/blog', 'Blog', 1, 3, datetime('now', '-100 days')),
(7, '/contact', 'Contact', 1, 2, datetime('now', '-100 days')),
(8, '/portfolio', 'Portfolio', 1, 6, datetime('now', '-80 days')),
(9, '/pricing', 'Pricing', 1, 7, datetime('now', '-70 days')),
(10, '/faq', 'FAQ', 1, 3, datetime('now', '-60 days')),
(11, '/privacy', 'Privacy Policy', 1, 2, datetime('now', '-50 days')),
(12, '/terms', 'Terms of Service', 1, 2, datetime('now', '-50 days'));

-- =============================================================================
-- 6. DATATYPES
-- =============================================================================
INSERT INTO datatypes (datatype_id, parent_id, label, type, author_id, date_created) VALUES
(1, NULL, 'Page', 'page', 1, datetime('now', '-120 days')),
(2, NULL, 'Article', 'article', 1, datetime('now', '-120 days')),
(3, NULL, 'Service', 'service', 1, datetime('now', '-120 days')),
(4, 2, 'Blog Post', 'blog_post', 1, datetime('now', '-110 days')),
(5, NULL, 'Contact Form', 'form', 1, datetime('now', '-110 days')),
(6, NULL, 'Portfolio Item', 'portfolio', 1, datetime('now', '-100 days')),
(7, NULL, 'FAQ Item', 'faq', 1, datetime('now', '-90 days')),
(8, NULL, 'Testimonial', 'testimonial', 1, datetime('now', '-80 days')),
(9, 1, 'Landing Page', 'landing', 1, datetime('now', '-70 days')),
(10, 2, 'News Article', 'news', 1, datetime('now', '-60 days'));

-- =============================================================================
-- 7. ADMIN_DATATYPES
-- =============================================================================
INSERT INTO admin_datatypes (admin_datatype_id, parent_id, label, type, author_id, date_created) VALUES
(1, NULL, 'Dashboard Widget', 'widget', 1, datetime('now', '-120 days')),
(2, NULL, 'User Profile', 'profile', 1, datetime('now', '-120 days')),
(3, NULL, 'Content Block', 'block', 1, datetime('now', '-120 days')),
(4, NULL, 'Media Item', 'media', 1, datetime('now', '-120 days')),
(5, 1, 'Stats Widget', 'stats_widget', 1, datetime('now', '-100 days')),
(6, 1, 'Chart Widget', 'chart_widget', 1, datetime('now', '-100 days')),
(7, NULL, 'System Setting', 'setting', 1, datetime('now', '-90 days')),
(8, NULL, 'Plugin Config', 'plugin_config', 1, datetime('now', '-80 days'));

-- =============================================================================
-- 8. ADMIN_FIELDS
-- =============================================================================
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created) VALUES
(1, 1, 'Widget Title', '', 'text', 1, datetime('now', '-120 days')),
(2, 1, 'Widget Type', 'stats', 'select', 1, datetime('now', '-120 days')),
(3, 2, 'Display Name', '', 'text', 1, datetime('now', '-120 days')),
(4, 2, 'Bio', '', 'textarea', 1, datetime('now', '-120 days')),
(5, 3, 'Block Title', '', 'text', 1, datetime('now', '-120 days')),
(6, 3, 'Block Content', '', 'richtext', 1, datetime('now', '-120 days')),
(7, 4, 'File Name', '', 'text', 1, datetime('now', '-120 days')),
(8, 4, 'Alt Text', '', 'text', 1, datetime('now', '-120 days')),
(9, 5, 'Metric Name', '', 'text', 1, datetime('now', '-100 days')),
(10, 5, 'Metric Value', '', 'number', 1, datetime('now', '-100 days')),
(11, 6, 'Chart Title', '', 'text', 1, datetime('now', '-100 days')),
(12, 6, 'Chart Data', '', 'json', 1, datetime('now', '-100 days')),
(13, 7, 'Setting Key', '', 'text', 1, datetime('now', '-90 days')),
(14, 7, 'Setting Value', '', 'text', 1, datetime('now', '-90 days')),
(15, 8, 'Plugin Name', '', 'text', 1, datetime('now', '-80 days')),
(16, 8, 'Plugin Enabled', 'false', 'boolean', 1, datetime('now', '-80 days'));

-- =============================================================================
-- 9. TOKENS
-- =============================================================================
INSERT INTO tokens (id, user_id, token_type, token, issued_at, expires_at, revoked) VALUES
(1, 2, 'access', 'tok_admin_access_example123', datetime('now', '-1 hour'), datetime('now', '+23 hours'), 0),
(2, 3, 'access', 'tok_editor_access_example456', datetime('now', '-2 hours'), datetime('now', '+22 hours'), 0),
(3, 2, 'refresh', 'tok_admin_refresh_example789', datetime('now', '-5 days'), datetime('now', '+25 days'), 0),
(4, 4, 'access', 'tok_contributor_access_abc', datetime('now', '-30 minutes'), datetime('now', '+23.5 hours'), 0),
(5, 6, 'access', 'tok_jsmith_access_xyz', datetime('now', '-10 minutes'), datetime('now', '+23.8 hours'), 0),
(6, 7, 'refresh', 'tok_mgarcia_refresh_def', datetime('now', '-3 days'), datetime('now', '+27 days'), 0),
(7, 8, 'access', 'tok_dwilson_access_ghi', datetime('now', '-45 minutes'), datetime('now', '+23.25 hours'), 0),
(8, 3, 'refresh', 'tok_editor_refresh_jkl', datetime('now', '-10 days'), datetime('now', '+20 days'), 0),
(9, 5, 'access', 'tok_viewer_access_mno', datetime('now', '-5 minutes'), datetime('now', '+23.9 hours'), 0),
(10, 10, 'access', 'tok_rbrown_access_pqr', datetime('now', '-15 minutes'), datetime('now', '+23.75 hours'), 0);

-- =============================================================================
-- 10. USER_OAUTH
-- =============================================================================
INSERT INTO user_oauth (user_oauth_id, user_id, oauth_provider, oauth_provider_user_id, access_token, refresh_token, token_expires_at, date_created) VALUES
(1, 2, 'google', 'google_user_12345', 'ya29.example_access_token_admin', '1//example_refresh_token_google_admin', datetime('now', '+1 hour'), datetime('now', '-180 days')),
(2, 3, 'github', 'github_user_67890', 'gho_example_github_token_editor', 'ghr_example_refresh_token_github_editor', datetime('now', '+1 hour'), datetime('now', '-120 days')),
(3, 6, 'google', 'google_user_54321', 'ya29.example_access_token_jsmith', '1//example_refresh_token_google_jsmith', datetime('now', '+1 hour'), datetime('now', '-45 days')),
(4, 7, 'github', 'github_user_98765', 'gho_example_github_token_mgarcia', 'ghr_example_refresh_token_github_mgarcia', datetime('now', '+1 hour'), datetime('now', '-30 days')),
(5, 4, 'microsoft', 'microsoft_user_11223', 'eyJ0eXAi_example_msft_token', 'RefreshToken_msft_example', datetime('now', '+1 hour'), datetime('now', '-90 days')),
(6, 8, 'google', 'google_user_33445', 'ya29.example_access_token_dwilson', '1//example_refresh_token_google_dwilson', datetime('now', '+1 hour'), datetime('now', '-20 days')),
(7, 9, 'github', 'github_user_55667', 'gho_example_github_token_sjohnson', 'ghr_example_refresh_token_github_sjohnson', datetime('now', '+1 hour'), datetime('now', '-15 days'));

-- =============================================================================
-- 10A. USER_SSH_KEYS
-- =============================================================================
INSERT INTO user_ssh_keys (ssh_key_id, user_id, public_key, key_type, fingerprint, label, date_created, last_used) VALUES
(1, 2, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8example...admin', 'ssh-rsa', 'SHA256:1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v', 'Admin Laptop', datetime('now', '-180 days'), datetime('now', '-1 hour')),
(2, 2, 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEXAMPLEadmin...desktop', 'ssh-ed25519', 'SHA256:9z8y7x6w5v4u3t2s1r0q9p8o7n6m5l4k3j2i1h0g9f8e', 'Admin Desktop', datetime('now', '-150 days'), datetime('now', '-2 days')),
(3, 3, 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEXAMPLEeditor...laptop', 'ssh-ed25519', 'SHA256:a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2', 'Editor MacBook', datetime('now', '-120 days'), datetime('now', '-30 minutes')),
(4, 6, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8example...jsmith', 'ssh-rsa', 'SHA256:z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4j3i2h1g0f9e8', 'Work Laptop', datetime('now', '-45 days'), datetime('now', '-3 hours')),
(5, 7, 'ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEXAMPLE...mgarcia', 'ecdsa-sha2-nistp256', 'SHA256:m1n2o3p4q5r6s7t8u9v0w1x2y3z4a5b6c7d8e9f0g1h2', 'Home Desktop', datetime('now', '-30 days'), datetime('now', '-5 days')),
(6, 10, 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEXAMPLErbrown...devbox', 'ssh-ed25519', 'SHA256:d1e2f3g4h5i6j7k8l9m0n1o2p3q4r5s6t7u8v9w0x1y2', 'Development Server', datetime('now', '-10 days'), datetime('now', '-1 day')),
(7, 10, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8example...rbrown_backup', 'ssh-rsa', 'SHA256:x1y2z3a4b5c6d7e8f9g0h1i2j3k4l5m6n7o8p9q0r1s2', 'Backup Laptop', datetime('now', '-8 days'), NULL);

-- =============================================================================
-- 11. TABLES
-- =============================================================================
INSERT INTO tables (id, label, author_id) VALUES
(1, 'permissions', 1),
(2, 'roles', 1),
(3, 'media_dimensions', 1),
(4, 'users', 1),
(5, 'admin_routes', 1),
(6, 'routes', 1),
(7, 'datatypes', 1),
(8, 'admin_datatypes', 1),
(9, 'admin_fields', 1),
(10, 'tokens', 1),
(11, 'user_oauth', 1),
(12, 'tables', 1),
(13, 'media', 1),
(14, 'sessions', 1),
(15, 'content_data', 1),
(16, 'content_fields', 1),
(17, 'fields', 1),
(18, 'admin_content_data', 1),
(19, 'admin_content_fields', 1),
(20, 'datatypes_fields', 1),
(21, 'admin_datatypes_fields', 1);

-- =============================================================================
-- 12. MEDIA
-- =============================================================================
INSERT INTO media (media_id, name, display_name, alt, caption, description, class, author_id, mimetype, dimensions, url, date_created) VALUES
(1, 'hero-banner.jpg', 'Hero Banner', 'Company hero banner image', 'Welcome to our site', 'Main hero banner for homepage', 'hero-image', 2, 'image/jpeg', '1920x1080', '/media/hero-banner.jpg', datetime('now', '-100 days')),
(2, 'about-team.jpg', 'Team Photo', 'Our amazing team', 'The team in 2025', 'Group photo of our team', 'team-photo', 2, 'image/jpeg', '1600x900', '/media/about-team.jpg', datetime('now', '-95 days')),
(3, 'service-web.jpg', 'Web Development', 'Web development services illustration', 'Custom web solutions', 'Web development service image', 'service-image', 3, 'image/jpeg', '800x600', '/media/service-web.jpg', datetime('now', '-90 days')),
(4, 'service-consulting.jpg', 'Consulting', 'Consulting services illustration', 'Expert consulting', 'Consulting service image', 'service-image', 3, 'image/jpeg', '800x600', '/media/service-consulting.jpg', datetime('now', '-90 days')),
(5, 'logo.svg', 'Company Logo', 'ModulaCMS logo', 'Our brand', 'Company logo in SVG format', 'logo', 1, 'image/svg+xml', '200x200', '/media/logo.svg', datetime('now', '-120 days')),
(6, 'portfolio-1.jpg', 'Project Alpha', 'E-commerce website screenshot', 'E-commerce Platform', 'Client project showcase', 'portfolio-image', 6, 'image/jpeg', '1920x1080', '/media/portfolio-1.jpg', datetime('now', '-80 days')),
(7, 'portfolio-2.jpg', 'Project Beta', 'Mobile app interface', 'Mobile App Design', 'Mobile application project', 'portfolio-image', 6, 'image/jpeg', '800x1200', '/media/portfolio-2.jpg', datetime('now', '-75 days')),
(8, 'blog-post-1.jpg', 'Blog Featured Image', 'Technology illustration', 'Modern Tech Stack', 'Blog post featured image', 'blog-image', 3, 'image/jpeg', '1200x630', '/media/blog-post-1.jpg', datetime('now', '-70 days')),
(9, 'testimonial-avatar-1.jpg', 'Client Avatar', 'Client photo', 'Happy Client', 'Testimonial client photo', 'avatar', 7, 'image/jpeg', '200x200', '/media/testimonial-avatar-1.jpg', datetime('now', '-60 days')),
(10, 'background-pattern.svg', 'Background Pattern', 'Decorative SVG pattern', 'Geometric Design', 'Repeatable background pattern', 'pattern', 2, 'image/svg+xml', '400x400', '/media/background-pattern.svg', datetime('now', '-50 days'));

-- =============================================================================
-- 13. SESSIONS
-- =============================================================================
INSERT INTO sessions (session_id, user_id, created_at, expires_at, last_access, ip_address, user_agent, session_data) VALUES
(1, 2, datetime('now', '-4 hours'), datetime('now', '+20 hours'), datetime('now', '-10 minutes'), '192.168.1.100', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', '{"preferences": {"theme": "dark", "language": "en"}}'),
(2, 3, datetime('now', '-2 hours'), datetime('now', '+22 hours'), datetime('now', '-5 minutes'), '192.168.1.101', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', '{"preferences": {"theme": "light", "language": "en"}}'),
(3, 6, datetime('now', '-1 hour'), datetime('now', '+23 hours'), datetime('now', '-2 minutes'), '192.168.1.102', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36', '{"preferences": {"theme": "dark", "language": "es"}}'),
(4, 7, datetime('now', '-30 minutes'), datetime('now', '+23.5 hours'), datetime('now', '-1 minute'), '192.168.1.103', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15', '{"preferences": {"theme": "light", "language": "en"}}'),
(5, 4, datetime('now', '-3 hours'), datetime('now', '+21 hours'), datetime('now', '-30 minutes'), '192.168.1.104', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0', '{"preferences": {"theme": "auto", "language": "en"}}'),
(6, 8, datetime('now', '-45 minutes'), datetime('now', '+23.25 hours'), datetime('now', '-3 minutes'), '192.168.1.105', 'Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/605.1.15', '{"preferences": {"theme": "light", "language": "en"}}'),
(7, 10, datetime('now', '-15 minutes'), datetime('now', '+23.75 hours'), datetime('now', '-1 minute'), '192.168.1.106', 'Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36', '{"preferences": {"theme": "dark", "language": "en"}}');

-- =============================================================================
-- 14. CONTENT_DATA (with Tree Structure)
-- =============================================================================
-- Demonstrating tree structure:
-- Home (1)
--   ├── About Us (2)
--   ├── Services (3)
--   │   ├── Web Development (5)
--   │   └── Consulting (6)
--   ├── Blog (7)
--   │   ├── First Post (8)
--   │   ├── Second Post (9)
--   │   └── Third Post (10)
--   ├── Portfolio (11)
--   │   ├── Project Alpha (12)
--   │   └── Project Beta (13)
--   ├── Pricing (14)
--   ├── FAQ (15)
--   └── Contact (4)

INSERT INTO content_data (content_data_id, parent_id, first_child_id, next_sibling_id, prev_sibling_id, route_id, datatype_id, author_id, date_created) VALUES
-- Root: Home (has children, no parent)
(1, NULL, 2, NULL, NULL, 1, 1, 2, datetime('now', '-120 days')),
-- Level 1: About Us (parent=1, next sibling=3)
(2, 1, NULL, 3, NULL, 2, 1, 2, datetime('now', '-120 days')),
-- Level 1: Services (parent=1, has children, prev=2, next=7)
(3, 1, 5, 7, 2, 3, 1, 2, datetime('now', '-120 days')),
-- Level 2: Web Development (parent=3, next sibling=6)
(5, 3, NULL, 6, NULL, 4, 3, 3, datetime('now', '-110 days')),
-- Level 2: Consulting (parent=3, prev sibling=5)
(6, 3, NULL, NULL, 5, 5, 3, 3, datetime('now', '-110 days')),
-- Level 1: Blog (parent=1, has children, prev=3, next=11)
(7, 1, 8, 11, 3, 6, 1, 3, datetime('now', '-100 days')),
-- Level 2: First Post (parent=7, next sibling=9)
(8, 7, NULL, 9, NULL, 6, 4, 3, datetime('now', '-90 days')),
-- Level 2: Second Post (parent=7, prev=8, next=10)
(9, 7, NULL, 10, 8, 6, 4, 3, datetime('now', '-80 days')),
-- Level 2: Third Post (parent=7, prev=9)
(10, 7, NULL, NULL, 9, 6, 4, 3, datetime('now', '-70 days')),
-- Level 1: Portfolio (parent=1, has children, prev=7, next=14)
(11, 1, 12, 14, 7, 8, 1, 6, datetime('now', '-80 days')),
-- Level 2: Project Alpha (parent=11, next=13)
(12, 11, NULL, 13, NULL, 8, 6, 6, datetime('now', '-80 days')),
-- Level 2: Project Beta (parent=11, prev=12)
(13, 11, NULL, NULL, 12, 8, 6, 6, datetime('now', '-75 days')),
-- Level 1: Pricing (parent=1, prev=11, next=15)
(14, 1, NULL, 15, 11, 9, 1, 7, datetime('now', '-70 days')),
-- Level 1: FAQ (parent=1, prev=14, next=4)
(15, 1, NULL, 4, 14, 10, 1, 3, datetime('now', '-60 days')),
-- Level 1: Contact (parent=1, prev sibling=15)
(4, 1, NULL, NULL, 15, 7, 5, 2, datetime('now', '-100 days'));

-- =============================================================================
-- 15. CONTENT_FIELDS
-- =============================================================================
INSERT INTO content_fields (content_field_id, route_id, content_data_id, field_id, field_value, author_id, date_created) VALUES
-- Home page fields
(1, 1, 1, 1, 'Welcome to ModulaCMS', 2, datetime('now', '-120 days')),
(2, 1, 1, 2, '<h1>Build Better Websites</h1><p>ModulaCMS is a modern headless CMS built with Go.</p>', 2, datetime('now', '-120 days')),
(3, 1, 1, 3, 'ModulaCMS - Modern headless CMS for agencies', 2, datetime('now', '-120 days')),
-- About Us fields
(4, 2, 2, 1, 'About Our Company', 2, datetime('now', '-120 days')),
(5, 2, 2, 2, '<p>We are a team of passionate developers building the next generation of content management.</p>', 2, datetime('now', '-120 days')),
(6, 2, 2, 3, 'Learn about our company and mission', 2, datetime('now', '-120 days')),
-- Services parent fields
(7, 3, 3, 1, 'Our Services', 2, datetime('now', '-120 days')),
(8, 3, 3, 2, '<p>We offer comprehensive web development and consulting services.</p>', 2, datetime('now', '-120 days')),
-- Web Development service
(9, 4, 5, 8, 'Web Development', 3, datetime('now', '-110 days')),
(10, 4, 5, 9, '<p>Custom web applications built with modern technologies.</p>', 3, datetime('now', '-110 days')),
(11, 4, 5, 10, '5000', 3, datetime('now', '-110 days')),
-- Consulting service
(12, 5, 6, 8, 'Technical Consulting', 3, datetime('now', '-110 days')),
(13, 5, 6, 9, '<p>Expert guidance on architecture and technology decisions.</p>', 3, datetime('now', '-110 days')),
(14, 5, 6, 10, '3000', 3, datetime('now', '-110 days')),
-- Blog parent
(15, 6, 7, 1, 'Blog', 3, datetime('now', '-100 days')),
(16, 6, 7, 2, '<p>Latest news and insights from our team.</p>', 3, datetime('now', '-100 days')),
-- Blog posts
(17, 6, 8, 4, 'Getting Started with ModulaCMS', 3, datetime('now', '-90 days')),
(18, 6, 8, 5, '<p>Learn how to build your first site with ModulaCMS...</p>', 3, datetime('now', '-90 days')),
(19, 6, 8, 6, 'Jane Editor', 3, datetime('now', '-90 days')),
(20, 6, 8, 7, '2025-01-10', 3, datetime('now', '-90 days')),
(21, 6, 9, 4, 'Advanced Tree Structures', 3, datetime('now', '-80 days')),
(22, 6, 9, 5, '<p>Deep dive into sibling-pointer trees...</p>', 3, datetime('now', '-80 days')),
(23, 6, 9, 6, 'Jane Editor', 3, datetime('now', '-80 days')),
(24, 6, 9, 7, '2025-01-12', 3, datetime('now', '-80 days')),
(25, 6, 10, 4, 'Performance Optimization Tips', 3, datetime('now', '-70 days')),
(26, 6, 10, 5, '<p>Best practices for optimizing your ModulaCMS site...</p>', 3, datetime('now', '-70 days')),
(27, 6, 10, 6, 'Bob Writer', 4, datetime('now', '-70 days')),
(28, 6, 10, 7, '2025-01-14', 4, datetime('now', '-70 days')),
-- Portfolio
(29, 8, 11, 1, 'Our Portfolio', 6, datetime('now', '-80 days')),
(30, 8, 11, 2, '<p>Showcasing our best work.</p>', 6, datetime('now', '-80 days')),
-- Pricing
(31, 9, 14, 1, 'Pricing Plans', 7, datetime('now', '-70 days')),
(32, 9, 14, 2, '<p>Choose the plan that works for you.</p>', 7, datetime('now', '-70 days')),
-- FAQ
(33, 10, 15, 1, 'Frequently Asked Questions', 3, datetime('now', '-60 days')),
(34, 10, 15, 2, '<p>Common questions and answers.</p>', 3, datetime('now', '-60 days')),
-- Contact
(35, 7, 4, 1, 'Get In Touch', 2, datetime('now', '-100 days')),
(36, 7, 4, 2, '<p>Contact us for more information.</p>', 2, datetime('now', '-100 days'));

-- =============================================================================
-- 16. FIELDS
-- =============================================================================
INSERT INTO fields (field_id, parent_id, label, data, type, author_id, date_created) VALUES
(1, 1, 'Page Title', '', 'text', 1, datetime('now', '-120 days')),
(2, 1, 'Page Content', '', 'richtext', 1, datetime('now', '-120 days')),
(3, 1, 'Meta Description', '', 'textarea', 1, datetime('now', '-120 days')),
(4, 2, 'Article Title', '', 'text', 1, datetime('now', '-120 days')),
(5, 2, 'Article Body', '', 'richtext', 1, datetime('now', '-120 days')),
(6, 2, 'Author Name', '', 'text', 1, datetime('now', '-120 days')),
(7, 2, 'Published Date', '', 'date', 1, datetime('now', '-120 days')),
(8, 3, 'Service Name', '', 'text', 1, datetime('now', '-120 days')),
(9, 3, 'Service Description', '', 'richtext', 1, datetime('now', '-120 days')),
(10, 3, 'Price', '', 'number', 1, datetime('now', '-120 days')),
(11, 5, 'Contact Email', '', 'email', 1, datetime('now', '-110 days')),
(12, 5, 'Message', '', 'textarea', 1, datetime('now', '-110 days')),
(13, 6, 'Project Name', '', 'text', 1, datetime('now', '-100 days')),
(14, 6, 'Project Description', '', 'richtext', 1, datetime('now', '-100 days')),
(15, 6, 'Project URL', '', 'url', 1, datetime('now', '-100 days')),
(16, 7, 'Question', '', 'text', 1, datetime('now', '-90 days')),
(17, 7, 'Answer', '', 'richtext', 1, datetime('now', '-90 days')),
(18, 8, 'Client Name', '', 'text', 1, datetime('now', '-80 days')),
(19, 8, 'Testimonial Text', '', 'textarea', 1, datetime('now', '-80 days')),
(20, 8, 'Rating', '5', 'number', 1, datetime('now', '-80 days'));

-- =============================================================================
-- 17. ADMIN_CONTENT_DATA (with Tree Structure)
-- =============================================================================
-- Admin dashboard structure:
-- Dashboard (1)
--   ├── Stats Widget (2)
--   ├── Activity Widget (3)
--   └── Quick Actions Widget (7)
-- User Management (4)
--   ├── User List (5)
--   └── Role Management (6)
-- Analytics (8)
--   ├── Traffic Chart (9)
--   └── Conversion Stats (10)

INSERT INTO admin_content_data (admin_content_data_id, parent_id, first_child_id, next_sibling_id, prev_sibling_id, admin_route_id, admin_datatype_id, author_id, date_created) VALUES
-- Root: Dashboard (has children)
(1, NULL, 2, NULL, NULL, 1, 1, 1, datetime('now', '-180 days')),
-- Level 1: Stats Widget (parent=1, next=3)
(2, 1, NULL, 3, NULL, 1, 5, 1, datetime('now', '-180 days')),
-- Level 1: Activity Widget (parent=1, prev=2, next=7)
(3, 1, NULL, 7, 2, 1, 1, 1, datetime('now', '-180 days')),
-- Level 1: Quick Actions Widget (parent=1, prev=3)
(7, 1, NULL, NULL, 3, 1, 1, 1, datetime('now', '-150 days')),
-- Root: User Management (has children)
(4, NULL, 5, NULL, NULL, 2, 2, 1, datetime('now', '-180 days')),
-- Level 1: User List (parent=4, next=6)
(5, 4, NULL, 6, NULL, 2, 2, 1, datetime('now', '-180 days')),
-- Level 1: Role Management (parent=4, prev=5)
(6, 4, NULL, NULL, 5, 2, 2, 1, datetime('now', '-180 days')),
-- Root: Analytics (has children)
(8, NULL, 9, NULL, NULL, 6, 6, 2, datetime('now', '-90 days')),
-- Level 1: Traffic Chart (parent=8, next=10)
(9, 8, NULL, 10, NULL, 6, 6, 2, datetime('now', '-90 days')),
-- Level 1: Conversion Stats (parent=8, prev=9)
(10, 8, NULL, NULL, 9, 6, 5, 2, datetime('now', '-90 days'));

-- =============================================================================
-- 18. ADMIN_CONTENT_FIELDS
-- =============================================================================
INSERT INTO admin_content_fields (admin_content_field_id, admin_route_id, admin_content_data_id, admin_field_id, admin_field_value, author_id, date_created) VALUES
-- Dashboard widgets
(1, 1, 2, 9, 'Total Users', 1, datetime('now', '-180 days')),
(2, 1, 2, 10, '1250', 1, datetime('now', '-180 days')),
(3, 1, 3, 1, 'Recent Activity', 1, datetime('now', '-180 days')),
(4, 1, 3, 2, 'activity', 1, datetime('now', '-180 days')),
(5, 1, 7, 1, 'Quick Actions', 1, datetime('now', '-150 days')),
(6, 1, 7, 2, 'actions', 1, datetime('now', '-150 days')),
-- User management
(7, 2, 5, 3, 'All Users', 1, datetime('now', '-180 days')),
(8, 2, 5, 4, 'Comprehensive list of all system users', 1, datetime('now', '-180 days')),
(9, 2, 6, 3, 'Role Management', 1, datetime('now', '-180 days')),
(10, 2, 6, 4, 'Manage user roles and permissions', 1, datetime('now', '-180 days')),
-- Analytics
(11, 6, 9, 11, 'Monthly Traffic', 2, datetime('now', '-90 days')),
(12, 6, 9, 12, '{"jan": 1200, "feb": 1350, "mar": 1500}', 2, datetime('now', '-90 days')),
(13, 6, 10, 9, 'Conversion Rate', 2, datetime('now', '-90 days')),
(14, 6, 10, 10, '3.5', 2, datetime('now', '-90 days'));

-- =============================================================================
-- 19. DATATYPES_FIELDS (Junction Table)
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
(16, 5, 12),
-- Portfolio Item uses project fields
(17, 6, 13),
(18, 6, 14),
(19, 6, 15),
-- FAQ Item uses question/answer fields
(20, 7, 16),
(21, 7, 17),
-- Testimonial uses client fields
(22, 8, 18),
(23, 8, 19),
(24, 8, 20),
-- Landing Page inherits from Page
(25, 9, 1),
(26, 9, 2),
(27, 9, 3),
-- News Article inherits from Article
(28, 10, 4),
(29, 10, 5),
(30, 10, 6),
(31, 10, 7);

-- =============================================================================
-- 20. ADMIN_DATATYPES_FIELDS (Junction Table)
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
(8, 4, 8),
-- Stats Widget uses metric fields
(9, 5, 9),
(10, 5, 10),
-- Chart Widget uses chart fields
(11, 6, 11),
(12, 6, 12),
-- System Setting uses setting fields
(13, 7, 13),
(14, 7, 14),
-- Plugin Config uses plugin fields
(15, 8, 15),
(16, 8, 16);

-- =============================================================================
-- DEMO DATA SUMMARY
-- =============================================================================
-- Total records inserted (all tables have multiple rows):
-- - Permissions: 20 rows
-- - Roles: 7 rows
-- - Media Dimensions: 10 rows
-- - Users: 10 rows
-- - Admin Routes: 10 rows
-- - Routes: 12 rows
-- - Datatypes: 10 rows
-- - Admin Datatypes: 8 rows
-- - Admin Fields: 16 rows
-- - Tokens: 10 rows
-- - User OAuth: 7 rows
-- - Tables: 21 rows (complete list of all database tables)
-- - Media: 10 rows
-- - Sessions: 7 rows
-- - Content Data: 15 rows (demonstrating tree structure)
-- - Content Fields: 36 rows
-- - Fields: 20 rows
-- - Admin Content Data: 10 rows (demonstrating tree structure)
-- - Admin Content Fields: 14 rows
-- - Datatypes Fields: 31 rows
-- - Admin Datatypes Fields: 16 rows
--
-- Order follows sql/create_order.md for proper foreign key dependencies
--
-- Tree Structure Examples:
-- 1. Frontend Content Tree (content_data):
--    - Multi-level hierarchy with sibling pointers
--    - Demonstrates parent-child and sibling relationships
--    - Root node (Home) with multiple children
--    - Branch nodes (Services, Blog, Portfolio) with sub-items
--
-- 2. Admin Content Tree (admin_content_data):
--    - Dashboard widgets organization
--    - User management hierarchy
--    - Analytics structure
--    - Demonstrates administrative content structure
