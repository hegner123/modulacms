package main

import (
	"fmt"
	"os"
	"strings"
)

type replacement struct {
	old string
	new string
}

func main() {
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
		os.Exit(1)
	}
	content := string(data)

	replacements := []replacement{
		// ===== SQLite Bootstrap Header =====
		{
			old: `func (d Database) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)
	permission := d.CreatePermission(CreatePermissionParams{`,
			new: `func (d Database) CreateBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)
	permission, err := d.CreatePermission(ctx, ac, CreatePermissionParams{`,
		},
		// Permission zero check
		{
			old: `	})
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole := d.CreateRole(CreateRoleParams{`,
			new: `	})
	if err != nil {
		return fmt.Errorf("failed to create system admin permission: %w", err)
	}
	if permission.PermissionID.IsZero() {
		return fmt.Errorf("failed to create system admin permission")
	}

	// 2. Create system admin role (role_id = 1)
	adminRole, err := d.CreateRole(ctx, ac, CreateRoleParams{`,
		},
		// Admin role zero check -> viewer role
		{
			old: `	})
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole := d.CreateRole(CreateRoleParams{`,
			new: `	})
	if err != nil {
		return fmt.Errorf("failed to create system admin role: %w", err)
	}
	if adminRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create system admin role")
	}

	// 3. Create viewer role (role_id = 4)
	viewerRole, err := d.CreateRole(ctx, ac, CreateRoleParams{`,
		},
		// Viewer role zero check -> system user
		{
			old: `	})
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(CreateUserParams{`,
			new: `	})
	if err != nil {
		return fmt.Errorf("failed to create viewer role: %w", err)
	}
	if viewerRole.RoleID.IsZero() {
		return fmt.Errorf("failed to create viewer role")
	}

	// 4. Create system admin user (user_id = 1)
	systemUser, err := d.CreateUser(ctx, ac, CreateUserParams{`,
		},
		// System user -> home route
		{
			old: `	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute := d.CreateRoute(CreateRouteParams{`,
			new: `	// 5. Create default home route (route_id = 1) - Recommended
	homeRoute, err := d.CreateRoute(ctx, ac, CreateRouteParams{`,
		},
		// Home route zero check -> page datatype
		{
			old: `	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype := d.CreateDatatype(CreateDatatypeParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default home route: %w", err)
	}
	if homeRoute.RouteID.IsZero() {
		return fmt.Errorf("failed to create default home route")
	}

	// 6. Create default page datatype (datatype_id = 1)
	pageDatatype, err := d.CreateDatatype(ctx, ac, CreateDatatypeParams{`,
		},
		// Page datatype zero check -> admin route
		{
			old: `	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute := d.CreateAdminRoute(CreateAdminRouteParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default page datatype: %w", err)
	}
	if pageDatatype.DatatypeID.IsZero() {
		return fmt.Errorf("failed to create default page datatype")
	}

	// 7. Create default admin route (admin_route_id = 1)
	adminRoute, err := d.CreateAdminRoute(ctx, ac, CreateAdminRouteParams{`,
		},
		// Admin route zero check -> admin datatype
		{
			old: `	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype := d.CreateAdminDatatype(CreateAdminDatatypeParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin route: %w", err)
	}
	if adminRoute.AdminRouteID.IsZero() {
		return fmt.Errorf("failed to create default admin route")
	}

	// 8. Create default admin datatype (admin_datatype_id = 1)
	adminDatatype, err := d.CreateAdminDatatype(ctx, ac, CreateAdminDatatypeParams{`,
		},
		// Admin datatype zero check -> admin field
		{
			old: `	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField := d.CreateAdminField(CreateAdminFieldParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin datatype: %w", err)
	}
	if adminDatatype.AdminDatatypeID.IsZero() {
		return fmt.Errorf("failed to create default admin datatype")
	}

	// 9. Create default admin field (admin_field_id = 1)
	adminField, err := d.CreateAdminField(ctx, ac, CreateAdminFieldParams{`,
		},
		// Admin field zero check -> field
		{
			old: `	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field := d.CreateField(CreateFieldParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin field: %w", err)
	}
	if adminField.AdminFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin field")
	}

	// 10. Create default field (field_id = 1)
	field, err := d.CreateField(ctx, ac, CreateFieldParams{`,
		},
		// Field zero check -> content_data
		{
			old: `	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData := d.CreateContentData(CreateContentDataParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default field: %w", err)
	}
	if field.FieldID.IsZero() {
		return fmt.Errorf("failed to create default field")
	}

	// 11. Create default content_data record (content_data_id = 1)
	contentData, err := d.CreateContentData(ctx, ac, CreateContentDataParams{`,
		},
		// Content data zero check -> admin_content_data
		{
			old: `	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData := d.CreateAdminContentData(CreateAdminContentDataParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default content_data: %w", err)
	}
	if contentData.ContentDataID.IsZero() {
		return fmt.Errorf("failed to create default content_data")
	}

	// 12. Create default admin_content_data record (admin_content_data_id = 1)
	adminContentData, err := d.CreateAdminContentData(ctx, ac, CreateAdminContentDataParams{`,
		},
		// Admin content data zero check -> content_field
		{
			old: `	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField := d.CreateContentField(CreateContentFieldParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin_content_data: %w", err)
	}
	if adminContentData.AdminContentDataID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_data")
	}

	// 13. Create default content_field (content_field_id = 1)
	contentField, err := d.CreateContentField(ctx, ac, CreateContentFieldParams{`,
		},
		// Content field zero check -> admin_content_field
		{
			old: `	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField := d.CreateAdminContentField(CreateAdminContentFieldParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default content_field: %w", err)
	}
	if contentField.ContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default content_field")
	}

	// 14. Create default admin_content_field (admin_content_field_id = 1)
	adminContentField, err := d.CreateAdminContentField(ctx, ac, CreateAdminContentFieldParams{`,
		},
		// Admin content field zero check -> media_dimension
		{
			old: `	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension := d.CreateMediaDimension(CreateMediaDimensionParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin_content_field: %w", err)
	}
	if adminContentField.AdminContentFieldID.IsZero() {
		return fmt.Errorf("failed to create default admin_content_field")
	}

	// 15. Create default media_dimension (md_id = 1) - Validation record
	mediaDimension, err := d.CreateMediaDimension(ctx, ac, CreateMediaDimensionParams{`,
		},
		// Media dimension zero check -> media
		{
			old: `	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media := d.CreateMedia(CreateMediaParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default media_dimension: %w", err)
	}
	if mediaDimension.MdID == "" {
		return fmt.Errorf("failed to create default media_dimension")
	}

	// 16. Create default media record (media_id = 1) - Validation record
	media, err := d.CreateMedia(ctx, ac, CreateMediaParams{`,
		},
		// Media zero check -> token
		{
			old: `	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token := d.CreateToken(CreateTokenParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default media: %w", err)
	}
	if media.MediaID.IsZero() {
		return fmt.Errorf("failed to create default media")
	}

	// 17. Create default token (id = 1) - Validation record
	token, err := d.CreateToken(ctx, ac, CreateTokenParams{`,
		},
		// Token zero check -> session
		{
			old: `	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(CreateSessionParams{`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default token: %w", err)
	}
	if token.ID == "" {
		return fmt.Errorf("failed to create default token")
	}

	// 18. Create default session (session_id = 1) - Validation record
	session, err := d.CreateSession(ctx, ac, CreateSessionParams{`,
		},
		// Session -> user_oauth
		{
			old: `	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(CreateUserOauthParams{`,
			new: `	// 19. Create default user_oauth record (user_oauth_id = 1) - Validation record
	userOauth, err := d.CreateUserOauth(ctx, ac, CreateUserOauthParams{`,
		},
		// User oauth -> user ssh key
		{
			old: `	userSshKey, err := d.CreateUserSshKey(CreateUserSshKeyParams{`,
			new: `	userSshKey, err := d.CreateUserSshKey(ctx, ac, CreateUserSshKeyParams{`,
		},
		// Table registry loop
		{
			old: `		table := d.CreateTable(CreateTableParams{Label: tableName})
		if table.ID == "" {`,
			new: `		table, err := d.CreateTable(ctx, ac, CreateTableParams{Label: tableName})
		if err != nil {
			return fmt.Errorf("failed to register table in tables registry: %s: %w", tableName, err)
		}
		if table.ID == "" {`,
		},
		// Datatype field junction
		{
			old: `	datatypeField := d.CreateDatatypeField(CreateDatatypeFieldParams{`,
			new: `	datatypeField, err := d.CreateDatatypeField(ctx, ac, CreateDatatypeFieldParams{`,
		},
		// Datatype field zero check
		{
			old: `	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default datatypes_fields: %w", err)
	}
	if datatypeField.ID == "" {
		return fmt.Errorf("failed to create default datatypes_fields")
	}

	// 22. Create default admin_datatypes_fields junction record`,
		},
		// Admin datatype field junction
		{
			old: `	adminDatatypeField := d.CreateAdminDatatypeField(CreateAdminDatatypeFieldParams{`,
			new: `	adminDatatypeField, err := d.CreateAdminDatatypeField(ctx, ac, CreateAdminDatatypeFieldParams{`,
		},
		// Admin datatype field zero check (before logger)
		{
			old: `	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo(`,
			new: `	if err != nil {
		return fmt.Errorf("failed to create default admin_datatypes_fields: %w", err)
	}
	if adminDatatypeField.ID == "" {
		return fmt.Errorf("failed to create default admin_datatypes_fields")
	}

	utility.DefaultLogger.Finfo(`,
		},
	}

	for _, r := range replacements {
		count := strings.Count(content, r.old)
		if count == 0 {
			fmt.Fprintf(os.Stderr, "WARNING: pattern not found: %q...\n", r.old[:min(80, len(r.old))])
			continue
		}
		// Apply to ALL occurrences (3 bootstrap methods have same patterns)
		content = strings.ReplaceAll(content, r.old, r.new)
		fmt.Printf("  Applied %d replacement(s) for pattern: %s...\n", count, r.old[:min(60, len(r.old))])
	}

	// Now add the header lines for MySQL and PostgreSQL bootstrap methods
	content = strings.Replace(content,
		`func (d MysqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)`,
		`func (d MysqlDatabase) CreateBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)`, 1)

	content = strings.Replace(content,
		`func (d PsqlDatabase) CreateBootstrapData() error {
	// 1. Create system admin permission (permission_id = 1)`,
		`func (d PsqlDatabase) CreateBootstrapData() error {
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "bootstrap", "system")

	// 1. Create system admin permission (permission_id = 1)`, 1)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Println("Bootstrap methods updated successfully")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
