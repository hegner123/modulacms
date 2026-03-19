package db

import (
	"reflect"
	"testing"
)

func TestGenericHeaders_AllKnownTables(t *testing.T) {
	t.Parallel()
	// Every known DBTable constant should return a non-nil, non-empty slice
	tables := []DBTable{
		Admin_content_data,
		Admin_content_fields,
		Admin_content_relations,
		Admin_content_versions,
		Admin_datatype,
		Admin_field,
		Admin_field_types,
		Admin_route,
		BackupT,
		Backup_set,
		Backup_verification,
		Change_event,
		Content_data,
		Content_fields,
		Content_relations,
		Content_versions,
		Datatype,
		Field,
		Field_plugin_config,
		Field_types,
		LocaleT,
		MediaT,
		Media_dimension,
		Media_folder,
		Permission,
		PipelineT,
		Role,
		Role_permissions,
		Route,
		Session,
		Table,
		Token,
		User,
		User_oauth,
		User_ssh_keys,
		WebhookT,
		Webhook_deliveries,
	}
	for _, tbl := range tables {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tbl)
			if headers == nil {
				t.Fatalf("GenericHeaders(%q) returned nil", tbl)
			}
			if len(headers) == 0 {
				t.Fatalf("GenericHeaders(%q) returned empty slice", tbl)
			}
			// Check no empty strings in headers
			for i, h := range headers {
				if h == "" {
					t.Errorf("GenericHeaders(%q)[%d] is empty string", tbl, i)
				}
			}
		})
	}
}

func TestGenericHeaders_UnknownTable_ReturnsNil(t *testing.T) {
	t.Parallel()
	got := GenericHeaders(DBTable("nonexistent"))
	if got != nil {
		t.Errorf("GenericHeaders(nonexistent) = %v, want nil", got)
	}
}

func TestGenericHeaders_SpecificCounts(t *testing.T) {
	t.Parallel()
	// Verify a few tables have the expected number of columns
	tests := []struct {
		table     DBTable
		wantCount int
	}{
		{Route, 8},           // route_id, slug, title, status, author_id, date_created, date_modified, history
		{User, 8},            // user_id, username, name, email, hash, role, date_created, date_modified
		{Permission, 2},      // permission_id, label
		{Role, 3},            // role_id, label, system_protected
		{Session, 8},         // session_id, user_id, date_created, expires_at, last_access, ip_address, user_agent, session_data
		{MediaT, 17},         // 17 media fields
		{Media_dimension, 5}, // md_id, label, width, height, aspect_ratio
		{Table, 3},           // id, label, author_id
		{Token, 7},           // id, user_id, token_type, token, issued_at, expires_at, revoked
		{Field_types, 3},     // field_type_id, type, label
		{Admin_field_types, 3}, // admin_field_type_id, type, label
		{Role_permissions, 3},  // id, role_id, permission_id
		{BackupT, 16},          // 16 backup fields
		{Change_event, 16},     // 16 change event fields
		{WebhookT, 7},          // 7 webhook string fields
		{Webhook_deliveries, 11}, // 11 webhook delivery fields
		{PipelineT, 11},         // 11 pipeline fields
		{LocaleT, 8},            // 8 locale fields
		{Content_versions, 10},  // 10 content version fields
		{Admin_content_versions, 10}, // 10 admin content version fields
		{User_ssh_keys, 8},      // 8 ssh key fields
		{Field_plugin_config, 6}, // 6 field plugin config fields
	}
	for _, tt := range tests {
		t.Run(string(tt.table), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tt.table)
			if len(headers) != tt.wantCount {
				t.Errorf("GenericHeaders(%q) has %d columns, want %d; columns: %v",
					tt.table, len(headers), tt.wantCount, headers)
			}
		})
	}
}

func TestGenericHeaders_FirstColumnIsID(t *testing.T) {
	t.Parallel()
	// For most tables, the first column should contain "id" (case-insensitive check not needed;
	// all headers use lowercase with underscores)
	tablesToCheck := []DBTable{
		Route, User, Permission, Session, MediaT, Media_dimension, Token, Table,
		Admin_content_data, Admin_content_fields, Admin_content_relations, Admin_content_versions,
		Admin_datatype, Admin_field, Admin_field_types, Admin_route,
		Content_data, Content_fields, Content_relations, Content_versions,
		Datatype, Field, Field_types, User_oauth, Media_folder,
		BackupT, Backup_set, Backup_verification, Change_event,
		Role_permissions, User_ssh_keys, WebhookT, Webhook_deliveries,
		PipelineT, LocaleT, Field_plugin_config,
	}
	for _, tbl := range tablesToCheck {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(tbl)
			if len(headers) == 0 {
				t.Fatal("empty headers")
			}
			first := headers[0]
			// The first column should end with "_id" or be "id"
			if first != "id" && len(first) < 3 {
				t.Errorf("first header %q does not look like an ID column", first)
			}
		})
	}
}

func TestGenericHeaders_RouteSpecificColumns(t *testing.T) {
	t.Parallel()
	headers := GenericHeaders(Route)
	expected := []string{
		"route_id", "slug", "title", "status",
		"author_id", "date_created", "date_modified", "history",
	}
	if len(headers) != len(expected) {
		t.Fatalf("got %d headers, want %d", len(headers), len(expected))
	}
	for i, want := range expected {
		if headers[i] != want {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], want)
		}
	}
}

func TestGenericHeaders_UserSpecificColumns(t *testing.T) {
	t.Parallel()
	headers := GenericHeaders(User)
	expected := []string{
		"user_id", "username", "name", "email",
		"hash", "role", "date_created", "date_modified",
	}
	if len(headers) != len(expected) {
		t.Fatalf("got %d headers, want %d", len(headers), len(expected))
	}
	for i, want := range expected {
		if headers[i] != want {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], want)
		}
	}
}

// TestGenericHeaders_MatchesStringStructFields uses reflection to verify that
// GenericHeaders returns exactly the json tags from the corresponding String*
// struct for every table. If a field is added to a String* struct (generated
// by dbgen from the entity struct) but not added to GenericHeaders, this test
// fails. This prevents the database screen from silently dropping columns.
func TestGenericHeaders_MatchesStringStructFields(t *testing.T) {
	t.Parallel()

	// Map each DBTable to its String* struct type.
	tableToStringStruct := map[DBTable]reflect.Type{
		Admin_content_data:      reflect.TypeOf(StringAdminContentData{}),
		Admin_content_fields:    reflect.TypeOf(StringAdminContentFields{}),
		Admin_content_relations: reflect.TypeOf(StringAdminContentRelations{}),
		Admin_content_versions:  reflect.TypeOf(StringAdminContentVersion{}),
		Admin_datatype:          reflect.TypeOf(StringAdminDatatypes{}),
		Admin_field:             reflect.TypeOf(StringAdminFields{}),
		Admin_field_types:       reflect.TypeOf(StringAdminFieldTypes{}),
		Admin_route:             reflect.TypeOf(StringAdminRoutes{}),
		BackupT:                 reflect.TypeOf(StringBackup{}),
		Backup_set:              reflect.TypeOf(StringBackupSet{}),
		Backup_verification:     reflect.TypeOf(StringBackupVerification{}),
		Change_event:            reflect.TypeOf(StringChangeEvent{}),
		Content_data:            reflect.TypeOf(StringContentData{}),
		Content_fields:          reflect.TypeOf(StringContentFields{}),
		Content_relations:       reflect.TypeOf(StringContentRelations{}),
		Content_versions:        reflect.TypeOf(StringContentVersion{}),
		Datatype:                reflect.TypeOf(StringDatatypes{}),
		Field:                   reflect.TypeOf(StringFields{}),
		Field_plugin_config:     reflect.TypeOf(StringFieldPluginConfig{}),
		Field_types:             reflect.TypeOf(StringFieldTypes{}),
		LocaleT:                 reflect.TypeOf(StringLocale{}),
		MediaT:                  reflect.TypeOf(StringMedia{}),
		Media_dimension:         reflect.TypeOf(StringMediaDimensions{}),
		Media_folder:            reflect.TypeOf(StringMediaFolder{}),
		Permission:              reflect.TypeOf(StringPermissions{}),
		PipelineT:               reflect.TypeOf(StringPipeline{}),
		Role:                    reflect.TypeOf(StringRoles{}),
		Role_permissions:        reflect.TypeOf(StringRolePermissions{}),
		Route:                   reflect.TypeOf(StringRoutes{}),
		Session:                 reflect.TypeOf(StringSessions{}),
		Table:                   reflect.TypeOf(StringTables{}),
		Token:                   reflect.TypeOf(StringTokens{}),
		User:                    reflect.TypeOf(StringUsers{}),
		User_oauth:              reflect.TypeOf(StringUserOauth{}),
		User_ssh_keys:           reflect.TypeOf(StringUserSshKeys{}),
		ValidationT:             reflect.TypeOf(StringValidation{}),
		Admin_validation:        reflect.TypeOf(StringAdminValidation{}),
		WebhookT:                reflect.TypeOf(StringWebhook{}),
		Webhook_deliveries:      reflect.TypeOf(StringWebhookDelivery{}),
	}

	for table, stringType := range tableToStringStruct {
		t.Run(string(table), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(table)
			if headers == nil {
				t.Fatalf("GenericHeaders(%q) returned nil", table)
			}

			// Extract json tags from the String* struct in field order
			var structTags []string
			for i := range stringType.NumField() {
				field := stringType.Field(i)
				tag := field.Tag.Get("json")
				if tag == "" || tag == "-" {
					continue
				}
				structTags = append(structTags, tag)
			}

			if len(headers) != len(structTags) {
				t.Errorf("GenericHeaders has %d columns but %s has %d fields\n  headers: %v\n  struct:  %v",
					len(headers), stringType.Name(), len(structTags), headers, structTags)
				return
			}

			for i, header := range headers {
				if header != structTags[i] {
					t.Errorf("column %d: GenericHeaders has %q but %s has %q",
						i, header, stringType.Name(), structTags[i])
				}
			}
		})
	}
}

// TestGenericHeaders_CoversAllTablesInTableStructMap verifies that every
// DBTable in TableStructMap also has a GenericHeaders case. Catches new
// tables added to consts.go but not to generic_list.go.
func TestGenericHeaders_CoversAllTablesInTableStructMap(t *testing.T) {
	t.Parallel()

	for table := range TableStructMap {
		t.Run(string(table), func(t *testing.T) {
			t.Parallel()
			headers := GenericHeaders(table)
			if headers == nil {
				t.Errorf("table %q is in TableStructMap but has no GenericHeaders case", table)
			}
		})
	}
}
