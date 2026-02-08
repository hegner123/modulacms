package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Entity definitions for CUD method transformation
type entity struct {
	file     string
	name     string // e.g. "Route", "Role"
	mapName  string // e.g. "MapRoute", "MapRole" - what to call on result
	idType   string // e.g. "types.RouteID", "types.RoleID", "string"
	idField  string // e.g. "RouteID", "RoleID", "ID"
	table    string // e.g. "routes", "roles"
	// For Creates that currently return value (not pointer):
	createReturnsBare bool
	// For Update success messages:
	updateSuccessField string // e.g. "s.Label", "s.Username"
	// For entities with Slug_2 in UpdateParams:
	hasSlug2 bool
}

func main() {
	baseDir := os.Args[1] // e.g. /Users/home/Documents/Code/Go_dev/modulacms/internal/db

	entities := []entity{
		{file: "route.go", name: "Route", mapName: "MapRoute", idType: "types.RouteID", idField: "RouteID", table: "routes", createReturnsBare: true, updateSuccessField: `s.Slug`, hasSlug2: true},
		{file: "role.go", name: "Role", mapName: "MapRole", idType: "types.RoleID", idField: "RoleID", table: "roles", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "permission.go", name: "Permission", mapName: "MapPermission", idType: "types.PermissionID", idField: "PermissionID", table: "permissions", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "session.go", name: "Session", mapName: "MapSession", idType: "types.SessionID", idField: "SessionID", table: "sessions", createReturnsBare: false, updateSuccessField: `s.SessionID`},
		{file: "token.go", name: "Token", mapName: "MapToken", idType: "string", idField: "ID", table: "tokens", createReturnsBare: true, updateSuccessField: `s.ID`},
		{file: "table.go", name: "Table", mapName: "MapTable", idType: "string", idField: "ID", table: "tables", createReturnsBare: true, updateSuccessField: `s.ID`},
		{file: "field.go", name: "Field", mapName: "MapField", idType: "types.FieldID", idField: "FieldID", table: "fields", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "datatype.go", name: "Datatype", mapName: "MapDatatype", idType: "types.DatatypeID", idField: "DatatypeID", table: "datatypes", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "datatype_field.go", name: "DatatypeField", mapName: "MapDatatypeField", idType: "string", idField: "ID", table: "datatypes_fields", createReturnsBare: true, updateSuccessField: `s.ID`},
		{file: "content_data.go", name: "ContentData", mapName: "MapContentData", idType: "types.ContentID", idField: "ContentDataID", table: "content_data", createReturnsBare: true, updateSuccessField: `s.ContentDataID`},
		{file: "content_field.go", name: "ContentField", mapName: "MapContentField", idType: "types.ContentFieldID", idField: "ContentFieldID", table: "content_fields", createReturnsBare: true, updateSuccessField: `s.ContentFieldID`},
		{file: "media.go", name: "Media", mapName: "MapMedia", idType: "types.MediaID", idField: "MediaID", table: "media", createReturnsBare: true, updateSuccessField: `s.MediaID`},
		{file: "media_dimension.go", name: "MediaDimension", mapName: "MapMediaDimension", idType: "string", idField: "MdID", table: "media_dimensions", createReturnsBare: true, updateSuccessField: `s.MdID`},
		{file: "admin_content_data.go", name: "AdminContentData", mapName: "MapAdminContentData", idType: "types.AdminContentID", idField: "AdminContentDataID", table: "admin_content_data", createReturnsBare: true, updateSuccessField: `s.AdminContentDataID`},
		{file: "admin_content_field.go", name: "AdminContentField", mapName: "MapAdminContentField", idType: "types.AdminContentFieldID", idField: "AdminContentFieldID", table: "admin_content_fields", createReturnsBare: true, updateSuccessField: `s.AdminContentFieldID`},
		{file: "admin_datatype.go", name: "AdminDatatype", mapName: "MapAdminDatatype", idType: "types.AdminDatatypeID", idField: "AdminDatatypeID", table: "admin_datatypes", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "admin_datatype_field.go", name: "AdminDatatypeField", mapName: "MapAdminDatatypeField", idType: "string", idField: "ID", table: "admin_datatypes_fields", createReturnsBare: true, updateSuccessField: `s.ID`},
		{file: "admin_field.go", name: "AdminField", mapName: "MapAdminField", idType: "types.AdminFieldID", idField: "AdminFieldID", table: "admin_fields", createReturnsBare: true, updateSuccessField: `s.Label`},
		{file: "admin_route.go", name: "AdminRoute", mapName: "MapAdminRoute", idType: "types.AdminRouteID", idField: "AdminRouteID", table: "admin_routes", createReturnsBare: true, updateSuccessField: `s.Slug`},
		{file: "user_oauth.go", name: "UserOauth", mapName: "MapUserOauth", idType: "types.UserOauthID", idField: "UserOauthID", table: "user_oauth", createReturnsBare: false, updateSuccessField: `s.UserOauthID`},
		{file: "user_ssh_keys.go", name: "UserSshKey", mapName: "MapUserSshKeys", idType: "string", idField: "SshKeyID", table: "user_ssh_keys", createReturnsBare: false, updateSuccessField: ``},
	}

	for _, e := range entities {
		path := filepath.Join(baseDir, e.file)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		content := string(data)
		modified := false

		// Find and replace each CUD method for all 3 drivers
		for _, driver := range []string{"Database", "MysqlDatabase", "PsqlDatabase"} {
			content, modified = transformCreate(content, e, driver, modified)
			content, modified = transformUpdate(content, e, driver, modified)
			content, modified = transformDelete(content, e, driver, modified)
		}

		// Special: UpdateDatatypeFieldSortOrder
		if e.file == "datatype_field.go" {
			content, modified = transformDatatypeFieldSortOrder(content, modified)
		}

		// Special: UpdateUserSshKeyLabel
		if e.file == "user_ssh_keys.go" {
			content, modified = transformUserSshKeyLabel(content, modified)
		}

		if modified {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
				os.Exit(1)
			}
			fmt.Printf("Updated %s\n", e.file)
		} else {
			fmt.Printf("No changes needed for %s\n", e.file)
		}
	}
}

func transformCreate(content string, e entity, driver string, prevModified bool) (string, bool) {
	modified := prevModified

	// Determine wrapper type names
	wrapperType := getWrapperType(e)

	// For bare creates: func (d Driver) CreateXxx(s CreateXxxParams) WrapperType {
	// For ptr creates: func (d Driver) CreateXxx(s CreateXxxParams) (*WrapperType, error) {
	var oldSig string
	if e.createReturnsBare {
		oldSig = fmt.Sprintf("func (d %s) Create%s(s Create%sParams) %s {", driver, e.name, e.name, wrapperType)
	} else {
		oldSig = fmt.Sprintf("func (d %s) Create%s(s Create%sParams) (*%s, error) {", driver, e.name, e.name, wrapperType)
		if e.file == "user_ssh_keys.go" {
			oldSig = fmt.Sprintf("func (d %s) CreateUserSshKey(params CreateUserSshKeyParams) (*UserSshKeys, error) {", driver)
		}
	}

	newSig := fmt.Sprintf("func (d %s) Create%s(ctx context.Context, ac audited.AuditContext, s Create%sParams) (*%s, error) {", driver, e.name, e.name, wrapperType)
	if e.file == "user_ssh_keys.go" {
		newSig = fmt.Sprintf("func (d %s) CreateUserSshKey(ctx context.Context, ac audited.AuditContext, params CreateUserSshKeyParams) (*UserSshKeys, error) {", driver)
	}

	idx := strings.Index(content, oldSig)
	if idx == -1 {
		return content, modified
	}

	// Find end of function (next func or end of section)
	endIdx := findFuncEnd(content, idx)
	if endIdx == -1 {
		fmt.Fprintf(os.Stderr, "  WARNING: Could not find end of Create%s for %s\n", e.name, driver)
		return content, modified
	}

	var newBody string
	cmdFactoryName := fmt.Sprintf("New%sCmd", e.name)
	if e.file == "user_ssh_keys.go" {
		cmdFactoryName = "NewUserSshKeyCmd"
	}

	mapFuncName := e.mapName
	entityLower := strings.ToLower(e.name[:1]) + e.name[1:]
	paramName := "s"
	if e.file == "user_ssh_keys.go" {
		paramName = "params"
	}

	newBody = fmt.Sprintf(`%s
	cmd := d.%s(ctx, ac, %s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %%w", err)
	}
	r := d.%s(result)
	return &r, nil
}`, newSig, cmdFactoryName, paramName, entityLower, mapFuncName)

	content = content[:idx] + newBody + content[endIdx:]
	modified = true
	return content, modified
}

func transformUpdate(content string, e entity, driver string, prevModified bool) (string, bool) {
	modified := prevModified

	oldSig := fmt.Sprintf("func (d %s) Update%s(s Update%sParams) (*string, error) {", driver, e.name, e.name)
	newSig := fmt.Sprintf("func (d %s) Update%s(ctx context.Context, ac audited.AuditContext, s Update%sParams) (*string, error) {", driver, e.name, e.name)

	idx := strings.Index(content, oldSig)
	if idx == -1 {
		return content, modified
	}

	endIdx := findFuncEnd(content, idx)
	if endIdx == -1 {
		fmt.Fprintf(os.Stderr, "  WARNING: Could not find end of Update%s for %s\n", e.name, driver)
		return content, modified
	}

	entityLower := strings.ToLower(e.name[:1]) + e.name[1:]
	successField := e.updateSuccessField
	if successField == "" {
		successField = `"updated"`
	}

	newBody := fmt.Sprintf(`%s
	cmd := d.Update%sCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update %s: %%w", err)
	}
	msg := fmt.Sprintf("Successfully updated %%v\n", %s)
	return &msg, nil
}`, newSig, e.name, entityLower, successField)

	content = content[:idx] + newBody + content[endIdx:]
	modified = true
	return content, modified
}

func transformDelete(content string, e entity, driver string, prevModified bool) (string, bool) {
	modified := prevModified

	oldSig := fmt.Sprintf("func (d %s) Delete%s(id %s) error {", driver, e.name, e.idType)
	newSig := fmt.Sprintf("func (d %s) Delete%s(ctx context.Context, ac audited.AuditContext, id %s) error {", driver, e.name, e.idType)

	// Special cases for user_ssh_keys
	if e.file == "user_ssh_keys.go" {
		oldSig = fmt.Sprintf("func (d %s) DeleteUserSshKey(id string) error {", driver)
		newSig = fmt.Sprintf("func (d %s) DeleteUserSshKey(ctx context.Context, ac audited.AuditContext, id string) error {", driver)
	}

	idx := strings.Index(content, oldSig)
	if idx == -1 {
		return content, modified
	}

	endIdx := findFuncEnd(content, idx)
	if endIdx == -1 {
		fmt.Fprintf(os.Stderr, "  WARNING: Could not find end of Delete%s for %s\n", e.name, driver)
		return content, modified
	}

	cmdFactoryName := fmt.Sprintf("Delete%sCmd", e.name)
	if e.file == "user_ssh_keys.go" {
		cmdFactoryName = "DeleteUserSshKeyCmd"
	}

	newBody := fmt.Sprintf(`%s
	cmd := d.%s(ctx, ac, id)
	return audited.Delete(cmd)
}`, newSig, cmdFactoryName)

	content = content[:idx] + newBody + content[endIdx:]
	modified = true
	return content, modified
}

func transformDatatypeFieldSortOrder(content string, prevModified bool) (string, bool) {
	modified := prevModified

	// SQLite
	old1 := `func (d Database) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdb.UpdateDatatypeFieldSortOrderParams{
		SortOrder: sortOrder,
		ID:        id,
	})
}`
	new1 := `func (d Database) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old1) {
		content = strings.Replace(content, old1, new1, 1)
		modified = true
	}

	// MySQL
	old2 := `func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbm.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}`
	new2 := `func (d MysqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old2) {
		content = strings.Replace(content, old2, new2, 1)
		modified = true
	}

	// PostgreSQL
	old3 := `func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(id string, sortOrder int64) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateDatatypeFieldSortOrder(d.Context, mdbp.UpdateDatatypeFieldSortOrderParams{
		SortOrder: int32(sortOrder),
		ID:        id,
	})
}`
	new3 := `func (d PsqlDatabase) UpdateDatatypeFieldSortOrder(ctx context.Context, ac audited.AuditContext, id string, sortOrder int64) error {
	cmd := d.UpdateDatatypeFieldSortOrderCmd(ctx, ac, id, sortOrder)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old3) {
		content = strings.Replace(content, old3, new3, 1)
		modified = true
	}

	return content, modified
}

func transformUserSshKeyLabel(content string, prevModified bool) (string, bool) {
	modified := prevModified

	// SQLite
	old1 := `func (d Database) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdb.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}`
	new1 := `func (d Database) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old1) {
		content = strings.Replace(content, old1, new1, 1)
		modified = true
	}

	// MySQL
	old2 := `func (d MysqlDatabase) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbm.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}`
	new2 := `func (d MysqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old2) {
		content = strings.Replace(content, old2, new2, 1)
		modified = true
	}

	// PostgreSQL
	old3 := `func (d PsqlDatabase) UpdateUserSshKeyLabel(id string, label string) error {
	queries := mdbp.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbp.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}`
	new3 := `func (d PsqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}`

	if strings.Contains(content, old3) {
		content = strings.Replace(content, old3, new3, 1)
		modified = true
	}

	return content, modified
}

func getWrapperType(e entity) string {
	switch e.name {
	case "Route":
		return "Routes"
	case "Role":
		return "Roles"
	case "Permission":
		return "Permissions"
	case "Session":
		return "Sessions"
	case "Token":
		return "Tokens"
	case "Table":
		return "Tables"
	case "Field":
		return "Fields"
	case "Datatype":
		return "Datatypes"
	case "DatatypeField":
		return "DatatypeFields"
	case "ContentData":
		return "ContentData"
	case "ContentField":
		return "ContentFields"
	case "Media":
		return "Media"
	case "MediaDimension":
		return "MediaDimensions"
	case "AdminContentData":
		return "AdminContentData"
	case "AdminContentField":
		return "AdminContentFields"
	case "AdminDatatype":
		return "AdminDatatypes"
	case "AdminDatatypeField":
		return "AdminDatatypeFields"
	case "AdminField":
		return "AdminFields"
	case "AdminRoute":
		return "AdminRoutes"
	case "UserOauth":
		return "UserOauth"
	case "UserSshKey":
		return "UserSshKeys"
	default:
		return e.name
	}
}

// findFuncEnd finds the closing brace of a function starting at idx.
// It counts brace depth starting from the opening brace.
func findFuncEnd(content string, idx int) int {
	// Find the first opening brace after idx
	braceStart := strings.Index(content[idx:], "{")
	if braceStart == -1 {
		return -1
	}
	braceStart += idx

	depth := 0
	inString := false
	inRawString := false
	prevChar := byte(0)

	for i := braceStart; i < len(content); i++ {
		ch := content[i]

		if inRawString {
			if ch == '`' {
				inRawString = false
			}
			continue
		}

		if inString {
			if ch == '"' && prevChar != '\\' {
				inString = false
			}
			prevChar = ch
			continue
		}

		switch ch {
		case '`':
			inRawString = true
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
		prevChar = ch
	}
	return -1
}
