-- =============================================================================
-- Component Datatype Seed Data (SQLite) - Generated 2026-02-10
-- =============================================================================
-- Creates Page and Section datatypes (with their fields), plus individual
-- AdminDatatypes for each component type (replacing the generic "Component"
-- datatype with Config JSON blob), plus one showcase instance of every
-- component on a /admin/showcase page.
--
-- This seed is additive -- it does not modify existing data.
-- All IDs start at 100 to avoid collisions with existing data.
--
-- Usage: sqlite3 modula.db < seed_components.sql
-- =============================================================================

-- =============================================================================
-- ID Mapping (ULID -> INTEGER)
-- =============================================================================
-- DATATYPES:
--   Page:            01KG5WR1VPZG63WF3Z1HTZSKZ4 -> 100
--   Section:         01KG5WR1VPZG63WF3Z1M3TX6C3 -> 101
--   table:           02KHDT00000000000000000001  -> 102
--   detail:          02KHDT00000000000000000002  -> 103
--   action-item:     02KHDT00000000000000000003  -> 104
--   grid:            02KHDT00000000000000000004  -> 105
--   form:            02KHDT00000000000000000005  -> 106
--   form-field:      02KHDT00000000000000000006  -> 107
--   stat-card:       02KHDT00000000000000000007  -> 108
--   chart:           02KHDT00000000000000000008  -> 109
--   alert:           02KHDT00000000000000000009  -> 110
--   breadcrumb:      02KHDT00000000000000000010  -> 111
--   breadcrumb-item: 02KHDT00000000000000000011  -> 112
--   rich-text:       02KHDT00000000000000000012  -> 113
--
-- FIELDS (Page):
--   Title:       01KG5WR1VWX2VG6QNY4STMVNPV -> 100
--   Layout:      01KG5WR1VWX2VG6QNY4W1ZJ0KK -> 101
--   Icon:        01KG5WR1VWX2VG6QNY4YSDAA2Z -> 102
--   Description: 01KG5WR1VWX2VG6QNY4ZM7ZBB7 -> 103
--   Order:       01KG67GGKZGVA6CAVVN7DX51NB -> 104
--
-- FIELDS (Section):
--   Heading:     01KG5WR1VWX2VG6QNY51T2ZY98 -> 105
--   Columns:     01KG5WR1VWX2VG6QNY54QST04M -> 106
--
-- FIELDS (Component):  02KHFL000...01 through 54 -> 107 through 160
--
-- JUNCTION: 100 through 160
--
-- ROUTE (showcase): 02KHRT00000000000000000001 -> 100
--
-- CONTENT DATA: 02KHCD000...01 through 25 -> 100 through 124
--
-- CONTENT FIELDS: 02KHCF000...01 through 74 -> 100 through 173
--
-- AUTHOR: 01KG5SY7D47H1WT33JKAF177AF -> 2

BEGIN;

-- =============================================================================
-- 1a. Admin Datatypes - Page and Section (referenced by content_data)
-- =============================================================================
-- These datatypes are needed for the showcase page and section nodes.

INSERT INTO admin_datatypes (admin_datatype_id, parent_id, label, type, author_id, date_created, date_modified) VALUES
  (100, NULL, 'Page',    'page',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (101, NULL, 'Section', 'section', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 1b. Admin Datatypes (12 component types)
-- =============================================================================
-- Label must match componentSchemaMap keys in schemas.ts (lowercase).
-- The frontend resolveComponentNode() checks node.datatype.info.label.toLowerCase().
--
-- parent_id links child datatypes to their parent component:
--   action-item     -> detail
--   form-field      -> form
--   breadcrumb-item -> breadcrumb

INSERT INTO admin_datatypes (admin_datatype_id, parent_id, label, type, author_id, date_created, date_modified) VALUES
  (102, NULL, 'table',           'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (103, NULL, 'detail',          'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (104, 103,  'action-item',     'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (105, NULL, 'grid',            'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (106, NULL, 'form',            'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (107, 106,  'form-field',      'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (108, NULL, 'stat-card',       'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (109, NULL, 'chart',           'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (110, NULL, 'alert',           'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (111, NULL, 'breadcrumb',      'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (112, 111,  'breadcrumb-item', 'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (113, NULL, 'rich-text',       'component', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 2a. Admin Fields - Page fields (5)
-- =============================================================================

INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (100, 100, 'Title',       '', 'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (101, 100, 'Layout',      '', 'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (102, 100, 'Icon',        '', 'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (103, 100, 'Description', '', 'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (104, 100, 'Order',       '', 'number', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 2b. Admin Fields - Section fields (2)
-- =============================================================================

INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (105, 101, 'Heading', '', 'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (106, 101, 'Columns', '', 'number', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 2c. Admin Fields (54 fields across 12 component datatypes)
-- =============================================================================
-- Field labels are converted to camelCase config keys by toCamelCase() in tree-utils.ts.
-- Example: "PageSize" -> "pageSize", "Load Endpoint" -> "loadEndpoint"
--
-- Column mapping (SQLite):
--   admin_field_id, parent_id (datatype), label, data, type, author_id, dates
-- NOTE: no validation or ui_config columns in SQLite schema

-- ---- Table fields (5) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (107, 102, 'Label',      '',                                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (108, 102, 'Endpoint',   '',                                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (109, 102, 'Columns',    '{"hint":"JSON array of column slugs"}',                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (110, 102, 'PageSize',   '{"default":10}',                                             'number', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (111, 102, 'Searchable', '{"options":[{"label":"Yes","value":"true"},{"label":"No","value":"false"}]}', 'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Detail fields (1) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (112, 103, 'Label', '', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- ActionItem fields (4) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (113, 104, 'Label',   '',                                                                                'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (114, 104, 'Href',    '',                                                                                'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (115, 104, 'Variant', '{"options":["primary","secondary","outline","ghost","destructive"]}',              'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (116, 104, 'Icon',    '{"hint":"lucide icon name"}',                                                     'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Grid fields (4) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (117, 105, 'Label',   '',                             'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (118, 105, 'Columns', '{"default":1}',                'number', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (119, 105, 'Rows',    '{"hint":"CSS grid-template-rows"}', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (120, 105, 'Gap',     '{"hint":"Tailwind spacing"}',  'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Form fields (6) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (121, 106, 'Label',         '',                                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (122, 106, 'Endpoint',      '',                                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (123, 106, 'Method',        '{"options":["POST","PUT"]}',            'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (124, 106, 'SubmitLabel',   '',                                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (125, 106, 'LoadEndpoint',  '',                                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (126, 106, 'LoadId',        '',                                      'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- FormField fields (6) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (127, 107, 'Slug',         '',                                                                                                              'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (128, 107, 'FieldType',    '{"options":["text","number","date","select","richtext","image","media_picker","email","url","textarea","slug"]}', 'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (129, 107, 'FieldLabel',   '',                                                                                                              'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (130, 107, 'Required',     '{"options":[{"label":"Yes","value":"true"},{"label":"No","value":"false"}]}',                                    'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (131, 107, 'Options',      '{"hint":"JSON object for field config"}',                                                                       'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (132, 107, 'DefaultValue', '',                                                                                                              'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- StatCard fields (8) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (133, 108, 'Label',      '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (134, 108, 'Value',      '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (135, 108, 'Endpoint',   '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (136, 108, 'ValueField', '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (137, 108, 'Icon',       '{"hint":"lucide icon name"}',                'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (138, 108, 'Trend',      '{"options":["up","down","neutral"]}',        'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (139, 108, 'TrendValue', '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (140, 108, 'Href',       '',                                           'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Chart fields (7) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (141, 109, 'Label',     '',                                                     'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (142, 109, 'Endpoint',  '',                                                     'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (143, 109, 'ChartType', '{"options":["line","bar","pie","area","donut"]}',       'select', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (144, 109, 'XAxis',     '',                                                     'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (145, 109, 'YAxis',     '',                                                     'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (146, 109, 'Height',    '{"default":300}',                                      'number', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (147, 109, 'Colors',    '{"hint":"JSON array of hex colors"}',                  'text',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Alert fields (6) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (148, 110, 'Label',       '',                                                                                'text',     2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (149, 110, 'Message',     '',                                                                                'richtext', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (150, 110, 'Severity',    '{"options":["info","warning","error","success"]}',                                 'select',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (151, 110, 'Dismissible', '{"options":[{"label":"Yes","value":"true"},{"label":"No","value":"false"}]}',      'select',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (152, 110, 'Href',        '',                                                                                'text',     2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (153, 110, 'LinkLabel',   '',                                                                                'text',     2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- Breadcrumb fields (2) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (154, 111, 'Label',     '', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (155, 111, 'Separator', '', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- BreadcrumbItem fields (2) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (156, 112, 'ItemLabel', '', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (157, 112, 'Href',      '', 'text', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- ---- RichText fields (3) ----
INSERT INTO admin_fields (admin_field_id, parent_id, label, data, type, author_id, date_created, date_modified) VALUES
  (158, 113, 'Label',   '',                                          'text',     2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (159, 113, 'Content', '',                                          'richtext', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (160, 113, 'Format',  '{"options":["html","markdown","plain"]}',   'select',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 3. Admin Datatypes <-> Fields junction
-- =============================================================================
-- Page (100) -> 5 fields, Section (101) -> 2 fields, then 54 component fields = 61 total

INSERT INTO admin_datatypes_fields (id, admin_datatype_id, admin_field_id) VALUES
  -- Page (DT 100) -> 5 fields
  (100, 100, 100),
  (101, 100, 101),
  (102, 100, 102),
  (103, 100, 103),
  (104, 100, 104),
  -- Section (DT 101) -> 2 fields
  (105, 101, 105),
  (106, 101, 106),
  -- Table (DT 102) -> 5 fields
  (107, 102, 107),
  (108, 102, 108),
  (109, 102, 109),
  (110, 102, 110),
  (111, 102, 111),
  -- Detail (DT 103) -> 1 field
  (112, 103, 112),
  -- ActionItem (DT 104) -> 4 fields
  (113, 104, 113),
  (114, 104, 114),
  (115, 104, 115),
  (116, 104, 116),
  -- Grid (DT 105) -> 4 fields
  (117, 105, 117),
  (118, 105, 118),
  (119, 105, 119),
  (120, 105, 120),
  -- Form (DT 106) -> 6 fields
  (121, 106, 121),
  (122, 106, 122),
  (123, 106, 123),
  (124, 106, 124),
  (125, 106, 125),
  (126, 106, 126),
  -- FormField (DT 107) -> 6 fields
  (127, 107, 127),
  (128, 107, 128),
  (129, 107, 129),
  (130, 107, 130),
  (131, 107, 131),
  (132, 107, 132),
  -- StatCard (DT 108) -> 8 fields
  (133, 108, 133),
  (134, 108, 134),
  (135, 108, 135),
  (136, 108, 136),
  (137, 108, 137),
  (138, 108, 138),
  (139, 108, 139),
  (140, 108, 140),
  -- Chart (DT 109) -> 7 fields
  (141, 109, 141),
  (142, 109, 142),
  (143, 109, 143),
  (144, 109, 144),
  (145, 109, 145),
  (146, 109, 146),
  (147, 109, 147),
  -- Alert (DT 110) -> 6 fields
  (148, 110, 148),
  (149, 110, 149),
  (150, 110, 150),
  (151, 110, 151),
  (152, 110, 152),
  (153, 110, 153),
  -- Breadcrumb (DT 111) -> 2 fields
  (154, 111, 154),
  (155, 111, 155),
  -- BreadcrumbItem (DT 112) -> 2 fields
  (156, 112, 156),
  (157, 112, 157),
  -- RichText (DT 113) -> 3 fields
  (158, 113, 158),
  (159, 113, 159),
  (160, 113, 160);

-- =============================================================================
-- 4. Admin Route for showcase page
-- =============================================================================
INSERT INTO admin_routes (admin_route_id, slug, title, status, author_id, date_created, date_modified) VALUES
  (100, 'showcase', 'Component Showcase', 1, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

-- =============================================================================
-- 5. Admin Content Data - tree nodes (25 nodes)
-- =============================================================================
-- Tree structure:
--   CD100 Page "Component Showcase" (root)
--     CD101 Section "Statistics"
--       CD102 stat-card
--     CD103 Section "Data Display"
--       CD104 table
--     CD105 Section "Visualization"
--       CD106 chart
--     CD107 Section "Navigation"
--       CD108 breadcrumb
--         CD109 breadcrumb-item "Home"
--         CD110 breadcrumb-item "Showcase"
--     CD111 Section "Notifications"
--       CD112 alert
--     CD113 Section "Actions"
--       CD114 detail
--         CD115 action-item "Create New"
--         CD116 action-item "Export"
--     CD117 Section "Layout"
--       CD118 grid
--     CD119 Section "Forms"
--       CD120 form
--         CD121 form-field "username"
--         CD122 form-field "email"
--     CD123 Section "Content"
--       CD124 rich-text

-- Disable foreign keys for content_data inserts (self-referential FKs)
PRAGMA foreign_keys = OFF;

INSERT INTO admin_content_data (admin_content_data_id, parent_id, first_child_id, next_sibling_id, prev_sibling_id, admin_route_id, admin_datatype_id, author_id, date_created, date_modified) VALUES
  -- CD100: Page (root)
  (100, NULL, 101, NULL, NULL, 100, 100, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD101: Section "Statistics" (first child of Page)
  (101, 100, 102, 103, NULL, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD102: stat-card
  (102, 101, NULL, NULL, NULL, 100, 108, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD103: Section "Data Display"
  (103, 100, 104, 105, 101, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD104: table
  (104, 103, NULL, NULL, NULL, 100, 102, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD105: Section "Visualization"
  (105, 100, 106, 107, 103, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD106: chart
  (106, 105, NULL, NULL, NULL, 100, 109, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD107: Section "Navigation"
  (107, 100, 108, 111, 105, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD108: breadcrumb
  (108, 107, 109, NULL, NULL, 100, 111, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD109: breadcrumb-item "Home"
  (109, 108, NULL, 110, NULL, 100, 112, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD110: breadcrumb-item "Showcase"
  (110, 108, NULL, NULL, 109, 100, 112, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD111: Section "Notifications"
  (111, 100, 112, 113, 107, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD112: alert
  (112, 111, NULL, NULL, NULL, 100, 110, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD113: Section "Actions"
  (113, 100, 114, 117, 111, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD114: detail
  (114, 113, 115, NULL, NULL, 100, 103, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD115: action-item "Create New"
  (115, 114, NULL, 116, NULL, 100, 104, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD116: action-item "Export"
  (116, 114, NULL, NULL, 115, 100, 104, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD117: Section "Layout"
  (117, 100, 118, 119, 113, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD118: grid
  (118, 117, NULL, NULL, NULL, 100, 105, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD119: Section "Forms"
  (119, 100, 120, 123, 117, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD120: form
  (120, 119, 121, NULL, NULL, 100, 106, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD121: form-field "username"
  (121, 120, NULL, 122, NULL, 100, 107, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD122: form-field "email"
  (122, 120, NULL, NULL, 121, 100, 107, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD123: Section "Content"
  (123, 100, 124, NULL, 119, 100, 101, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  -- CD124: rich-text
  (124, 123, NULL, NULL, NULL, 100, 113, 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

PRAGMA foreign_keys = ON;

-- =============================================================================
-- 6. Admin Content Fields (field values for each content node)
-- =============================================================================
-- Format: content_field_id, route_id, content_data_id, field_id, field_value, author_id, dates
-- Field labels are matched by toCamelCase() -> schema key in the frontend.

INSERT INTO admin_content_fields (admin_content_field_id, admin_route_id, admin_content_data_id, admin_field_id, admin_field_value, author_id, date_created, date_modified) VALUES
  -- ---- CD100: Page "Component Showcase" ----
  (100, 100, 100, 100, 'Component Showcase',                   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (101, 100, 100, 101, 'sidebar',                              2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (102, 100, 100, 102, 'plugins',                              2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (103, 100, 100, 103, 'One instance of every component type', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (104, 100, 100, 104, '99',                                   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD101: Section "Statistics" ----
  (105, 100, 101, 105, 'Statistics',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (106, 100, 101, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD102: stat-card ----
  (107, 100, 102, 133, 'Total Users',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (108, 100, 102, 134, '1,234',         2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (109, 100, 102, 137, 'users',         2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (110, 100, 102, 138, 'up',            2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (111, 100, 102, 139, '+12.5%',        2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (112, 100, 102, 140, '/admin/users',  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD103: Section "Data Display" ----
  (113, 100, 103, 105, 'Data Display',  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (114, 100, 103, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD104: table ----
  (115, 100, 104, 107, 'Users',                                          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (116, 100, 104, 108, '/api/v1/users',                                   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (117, 100, 104, 109, '["id","username","email","role","created_at"]',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (118, 100, 104, 110, '10',                                              2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (119, 100, 104, 111, 'true',                                            2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD105: Section "Visualization" ----
  (120, 100, 105, 105, 'Visualization', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (121, 100, 105, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD106: chart ----
  (122, 100, 106, 141, 'User Signups',          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (123, 100, 106, 142, '/api/v1/stats/signups', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (124, 100, 106, 143, 'line',                  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (125, 100, 106, 144, 'date',                  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (126, 100, 106, 145, 'count',                 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (127, 100, 106, 146, '300',                   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD107: Section "Navigation" ----
  (128, 100, 107, 105, 'Navigation',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (129, 100, 107, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD108: breadcrumb ----
  (130, 100, 108, 154, 'Page Breadcrumb', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (131, 100, 108, 155, '/',                2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD109: breadcrumb-item "Home" ----
  (132, 100, 109, 156, 'Home',              2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (133, 100, 109, 157, '/admin/dashboard',  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD110: breadcrumb-item "Showcase" ----
  (134, 100, 110, 156, 'Showcase',          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD111: Section "Notifications" ----
  (135, 100, 111, 105, 'Notifications', 2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (136, 100, 111, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD112: alert ----
  (137, 100, 112, 148, 'System Notice',                                                  2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (138, 100, 112, 149, 'This is a showcase page demonstrating all component types.',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (139, 100, 112, 150, 'info',                                                          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (140, 100, 112, 151, 'true',                                                          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD113: Section "Actions" ----
  (141, 100, 113, 105, 'Actions',       2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (142, 100, 113, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD114: detail ----
  (143, 100, 114, 112, 'Quick Actions',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD115: action-item "Create New" ----
  (144, 100, 115, 113, 'Create New',       2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (145, 100, 115, 114, '/admin/content',   2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (146, 100, 115, 115, 'primary',          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD116: action-item "Export" ----
  (147, 100, 116, 113, 'Export Data',      2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (148, 100, 116, 115, 'outline',          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD117: Section "Layout" ----
  (149, 100, 117, 105, 'Layout',        2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (150, 100, 117, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD118: grid ----
  (151, 100, 118, 117, 'Example Grid',     2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (152, 100, 118, 118, '2',                2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (153, 100, 118, 119, 'auto',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (154, 100, 118, 120, '4',                2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD119: Section "Forms" ----
  (155, 100, 119, 105, 'Forms',         2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (156, 100, 119, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD120: form ----
  (157, 100, 120, 121, 'Create User',      2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (158, 100, 120, 122, '/api/v1/users',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (159, 100, 120, 123, 'POST',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (160, 100, 120, 124, 'Create User',      2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD121: form-field "username" ----
  (161, 100, 121, 127, 'username',         2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (162, 100, 121, 128, 'text',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (163, 100, 121, 129, 'Username',         2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (164, 100, 121, 130, 'true',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD122: form-field "email" ----
  (165, 100, 122, 127, 'email',            2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (166, 100, 122, 128, 'email',            2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (167, 100, 122, 129, 'Email Address',    2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (168, 100, 122, 130, 'true',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD123: Section "Content" ----
  (169, 100, 123, 105, 'Content',       2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (170, 100, 123, 106, '1',             2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),

  -- ---- CD124: rich-text ----
  (171, 100, 124, 158, 'Welcome Message',                                                                                                          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (172, 100, 124, 159, '<h3>Component Showcase</h3><p>This page demonstrates each admin component type rendered from the content tree.</p>',          2, '2026-02-10 00:00:00', '2026-02-10 00:00:00'),
  (173, 100, 124, 160, 'html',                                                                                                                       2, '2026-02-10 00:00:00', '2026-02-10 00:00:00');

COMMIT;
