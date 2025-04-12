-- Table 1: permissions
INSERT INTO permissions (table_id, mode, label) VALUES (1, 1, 'create_permission');
INSERT INTO permissions (table_id, mode, label) VALUES (1, 2, 'read_permission');
INSERT INTO permissions (table_id, mode, label) VALUES (1, 3, 'update_permission');
INSERT INTO permissions (table_id, mode, label) VALUES (1, 4, 'delete_permission');

-- Table 2: role
INSERT INTO permissions (table_id, mode, label) VALUES (2, 1, 'create_role');
INSERT INTO permissions (table_id, mode, label) VALUES (2, 2, 'read_role');
INSERT INTO permissions (table_id, mode, label) VALUES (2, 3, 'update_role');
INSERT INTO permissions (table_id, mode, label) VALUES (2, 4, 'delete_role');

-- Table 3: media_dimension
INSERT INTO permissions (table_id, mode, label) VALUES (3, 1, 'create_media_dimension');
INSERT INTO permissions (table_id, mode, label) VALUES (3, 2, 'read_media_dimension');
INSERT INTO permissions (table_id, mode, label) VALUES (3, 3, 'update_media_dimension');
INSERT INTO permissions (table_id, mode, label) VALUES (3, 4, 'delete_media_dimension');

-- Table 4: user
INSERT INTO permissions (table_id, mode, label) VALUES (4, 1, 'create_user');
INSERT INTO permissions (table_id, mode, label) VALUES (4, 2, 'read_user');
INSERT INTO permissions (table_id, mode, label) VALUES (4, 3, 'update_user');
INSERT INTO permissions (table_id, mode, label) VALUES (4, 4, 'delete_user');

-- Table 5: admin_route
INSERT INTO permissions (table_id, mode, label) VALUES (5, 1, 'create_admin_route');
INSERT INTO permissions (table_id, mode, label) VALUES (5, 2, 'read_admin_route');
INSERT INTO permissions (table_id, mode, label) VALUES (5, 3, 'update_admin_route');
INSERT INTO permissions (table_id, mode, label) VALUES (5, 4, 'delete_admin_route');

-- Table 6: route
INSERT INTO permissions (table_id, mode, label) VALUES (6, 1, 'create_route');
INSERT INTO permissions (table_id, mode, label) VALUES (6, 2, 'read_route');
INSERT INTO permissions (table_id, mode, label) VALUES (6, 3, 'update_route');
INSERT INTO permissions (table_id, mode, label) VALUES (6, 4, 'delete_route');

-- Table 7: datatype
INSERT INTO permissions (table_id, mode, label) VALUES (7, 1, 'create_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (7, 2, 'read_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (7, 3, 'update_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (7, 4, 'delete_datatype');

-- Table 8: field
INSERT INTO permissions (table_id, mode, label) VALUES (8, 1, 'create_field');
INSERT INTO permissions (table_id, mode, label) VALUES (8, 2, 'read_field');
INSERT INTO permissions (table_id, mode, label) VALUES (8, 3, 'update_field');
INSERT INTO permissions (table_id, mode, label) VALUES (8, 4, 'delete_field');

-- Table 9: admin_datatype
INSERT INTO permissions (table_id, mode, label) VALUES (9, 1, 'create_admin_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (9, 2, 'read_admin_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (9, 3, 'update_admin_datatype');
INSERT INTO permissions (table_id, mode, label) VALUES (9, 4, 'delete_admin_datatype');

-- Table 10: admin_field
INSERT INTO permissions (table_id, mode, label) VALUES (10, 1, 'create_admin_field');
INSERT INTO permissions (table_id, mode, label) VALUES (10, 2, 'read_admin_field');
INSERT INTO permissions (table_id, mode, label) VALUES (10, 3, 'update_admin_field');
INSERT INTO permissions (table_id, mode, label) VALUES (10, 4, 'delete_admin_field');

-- Table 11: token
INSERT INTO permissions (table_id, mode, label) VALUES (11, 1, 'create_token');
INSERT INTO permissions (table_id, mode, label) VALUES (11, 2, 'read_token');
INSERT INTO permissions (table_id, mode, label) VALUES (11, 3, 'update_token');
INSERT INTO permissions (table_id, mode, label) VALUES (11, 4, 'delete_token');

-- Table 12: user_oauth
INSERT INTO permissions (table_id, mode, label) VALUES (12, 1, 'create_user_oauth');
INSERT INTO permissions (table_id, mode, label) VALUES (12, 2, 'read_user_oauth');
INSERT INTO permissions (table_id, mode, label) VALUES (12, 3, 'update_user_oauth');
INSERT INTO permissions (table_id, mode, label) VALUES (12, 4, 'delete_user_oauth');

-- Table 13: table
INSERT INTO permissions (table_id, mode, label) VALUES (13, 1, 'create_table');
INSERT INTO permissions (table_id, mode, label) VALUES (13, 2, 'read_table');
INSERT INTO permissions (table_id, mode, label) VALUES (13, 3, 'update_table');
INSERT INTO permissions (table_id, mode, label) VALUES (13, 4, 'delete_table');

-- Table 14: media (using label '<mode>_media')
INSERT INTO permissions (table_id, mode, label) VALUES (14, 1, 'create_media');
INSERT INTO permissions (table_id, mode, label) VALUES (14, 2, 'read_media');
INSERT INTO permissions (table_id, mode, label) VALUES (14, 3, 'update_media');
INSERT INTO permissions (table_id, mode, label) VALUES (14, 4, 'delete_media');

-- Table 15: session
INSERT INTO permissions (table_id, mode, label) VALUES (15, 1, 'create_session');
INSERT INTO permissions (table_id, mode, label) VALUES (15, 2, 'read_session');
INSERT INTO permissions (table_id, mode, label) VALUES (15, 3, 'update_session');
INSERT INTO permissions (table_id, mode, label) VALUES (15, 4, 'delete_session');

-- Table 16: content_data
INSERT INTO permissions (table_id, mode, label) VALUES (16, 1, 'create_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (16, 2, 'read_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (16, 3, 'update_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (16, 4, 'delete_content_data');

-- Table 17: content_fields
INSERT INTO permissions (table_id, mode, label) VALUES (17, 1, 'create_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (17, 2, 'read_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (17, 3, 'update_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (17, 4, 'delete_content_fields');

-- Table 18: admin_content_data
INSERT INTO permissions (table_id, mode, label) VALUES (18, 1, 'create_admin_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (18, 2, 'read_admin_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (18, 3, 'update_admin_content_data');
INSERT INTO permissions (table_id, mode, label) VALUES (18, 4, 'delete_admin_content_data');

-- Table 19: admin_content_fields
INSERT INTO permissions (table_id, mode, label) VALUES (19, 1, 'create_admin_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (19, 2, 'read_admin_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (19, 3, 'update_admin_content_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (19, 4, 'delete_admin_content_fields');

-- Table 20: datatypes_fields
INSERT INTO permissions (table_id, mode, label) VALUES (20, 1, 'create_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (20, 2, 'read_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (20, 3, 'update_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (20, 4, 'delete_datatypes_fields');

-- Table 21: admin_datatypes_fields
INSERT INTO permissions (table_id, mode, label) VALUES (21, 1, 'create_admin_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (21, 2, 'read_admin_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (21, 3, 'update_admin_datatypes_fields');
INSERT INTO permissions (table_id, mode, label) VALUES (21, 4, 'delete_admin_datatypes_fields');

