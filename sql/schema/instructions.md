Create MySql Insert statements sequentially from the following list of tables and table definition. Each table should have 4 insert 
statements with the mode column containing the values 1 through 4 for 'create','read','update',and 'delete' modes. 
Use the preprepared insert statements as a template for the label column

tables:
1 permissions
2 role
3 media_dimension
4 user
5 admin_route
6 route
7 datatype
8 field
9 admin_datatype
10 admin_field
11 token
12 user_oauth
13 table
14 media
15 session
16 content_data
17 content_fields
18 admin_content_data
19 admin_content_fields
20 datatypes_fields
21 admin_datatypes_fields


```sql
CREATE TABLE IF NOT EXISTS permissions (
    permission_id INT AUTO_INCREMENT PRIMARY KEY,
    table_id INT NOT NULL,
    mode INT NOT NULL,
    label VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO modula_db.permissions (table_id, mode, label) VALUES (1, 1, 'create_permission');
INSERT INTO modula_db.permissions (table_id, mode, label) VALUES (1, 2, 'read_permission');
INSERT INTO modula_db.permissions (table_id, mode, label) VALUES (1, 3, 'update_permission');
INSERT INTO modula_db.permissions (table_id, mode, label) VALUES (1, 4, 'delete_permission');
```


