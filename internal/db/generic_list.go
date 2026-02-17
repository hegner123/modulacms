package db

// GenericHeaders returns the column header names for the given table.
func GenericHeaders(t DBTable) []string {
	switch t {
	case Admin_content_data:
		return []string{
			"admin_content_data_id",
			"parent_id",
			"admin_route_id",
			"admin_datatype_id",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_content_fields:
		return []string{
			"admin_content_field_id",
			"admin_route_id",
			"admin_content_data_id",
			"admin_field_id",
			"admin_field_value",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_datatype:
		return []string{
			"admin_datatype_id",
			"parent_id",
			"label",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_field:
		return []string{
			"admin_field_id",
			"parent_id",
			"label",
			"data",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_datatype_fields:
		return []string{
			"id",
			"admin_datatype_id",
			"admin_field_id",
		}
	case Admin_route:
		return []string{
			"admin_route_id",
			"slug",
			"title",
			"status",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Content_data:
		return []string{
			"content_data_id",
			"parent_id",
			"route_id",
			"datatype_id",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Content_fields:
		return []string{
			"content_field_id",
			"route_id",
			"content_data_id",
			"field_id",
			"field_value",
			"date_created",
			"date_modified",
			"history",
		}
	case Datatype_fields:
		return []string{
			"id",
			"datatype_id",
			"field_id",
		}
	case Datatype:
		return []string{
			"datatype_id",
			"parent_id",
			"label",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Field:
		return []string{
			"field_id",
			"parent_id",
			"label",
			"data",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case MediaT:
		return []string{
			"media_id",
			"name",
			"display_name",
			"alt",
			"caption",
			"description",
			"class",
			"mimetype",
			"dimensions",
			"url",
			"srcset",
			"author_id",
			"date_created",
			"date_modified",
		}
	case Media_dimension:
		return []string{
			"md_id",
			"label",
			"width",
			"height",
			"aspect_ratio",
		}
	case Permission:
		return []string{
			"permission_id",
			"table_id",
			"mode",
			"label",
		}
	case Role:
		return []string{
			"role_id",
			"label",
			"permissions",
		}
	case Route:
		return []string{
			"route_id",
			"slug",
			"title",
			"status",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Session:
		return []string{
			"session_id",
			"user_id",
			"created_at",
			"expires_at",
			"last_access",
			"ip_address",
			"user_agent",
			"session_data",
		}
	case Table:
		return []string{
			"id",
			"label",
			"author_id",
		}
	case Token:
		return []string{
			"id",
			"user_id",
			"token_type",
			"token",
			"issued_at",
			"expires_at",
			"revoked",
		}
	case User:
		return []string{
			"user_id",
			"username",
			"name",
			"email",
			"hash",
			"role",
			"date_created",
			"date_modified",
		}
	case User_oauth:
		return []string{
			"user_oauth_id",
			"user_id",
			"oauth_provider",
			"oauth_provider_user_id",
			"access_token",
			"refresh_token",
			"token_expires_at",
			"date_created",
		}
	}
	return nil
}

// GenericList retrieves all rows from the given table as string slices for TUI display.
func GenericList(t DBTable, d DbDriver) ([][]string, error) {
	switch t {
	case Admin_content_data:
		a, err := d.ListAdminContentData()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminContentData(row)
			r := []string{
				s.AdminContentDataID,
				s.ParentID,
				s.AdminRouteID,
				s.AdminDatatypeID,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_content_fields:
		a, err := d.ListAdminContentFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminContentField(row)
			r := []string{
				s.AdminContentFieldID,
				s.AdminRouteID,
				s.AdminContentDataID,
				s.AdminFieldID,
				s.AdminFieldValue,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_datatype:
		a, err := d.ListAdminDatatypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminDatatype(row)
			r := []string{
				s.AdminDatatypeID,
				s.ParentID,
				s.Label,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_field:
		a, err := d.ListAdminFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminField(row)
			r := []string{
				s.AdminFieldID,
				s.ParentID,
				s.Label,
				s.Data,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_datatype_fields:
		a, err := d.ListAdminDatatypeField()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminDatatypeField(row)
			r := []string{
				s.ID,
				s.AdminDatatypeID,
				s.AdminFieldID,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_route:
		a, err := d.ListAdminRoutes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminRoute(row)
			r := []string{
				s.AdminRouteID,
				s.Slug,
				s.Title,
				s.Status,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_data:
		a, err := d.ListContentData()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringContentData(row)
			r := []string{
				s.ContentDataID,
				s.ParentID,
				s.RouteID,
				s.DatatypeID,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_fields:
		a, err := d.ListContentFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringContentField(row)
			r := []string{
				s.ContentFieldID,
				s.RouteID,
				s.ContentDataID,
				s.FieldID,
				s.FieldValue,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Datatype_fields:
		a, err := d.ListDatatypeField()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringDatatypeField(row)
			r := []string{
				s.ID,
				s.DatatypeID,
				s.FieldID,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Datatype:
		a, err := d.ListDatatypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringDatatype(row)
			r := []string{
				s.DatatypeID,
				s.ParentID,
				s.Label,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Field:
		a, err := d.ListFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringField(row)
			r := []string{
				s.FieldID,
				s.ParentID,
				s.Label,
				s.Data,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case MediaT:
		a, err := d.ListMedia()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringMedia(row)
			r := []string{
				s.MediaID,
				s.Name,
				s.DisplayName,
				s.Alt,
				s.Caption,
				s.Description,
				s.Class,
				s.Mimetype,
				s.Dimensions,
				s.Url,
				s.Srcset,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Media_dimension:
		a, err := d.ListMediaDimensions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringMediaDimension(row)
			r := []string{
				s.MdID,
				s.Label,
				s.Width,
				s.Height,
				s.AspectRatio,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Permission:
		a, err := d.ListPermissions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringPermission(row)
			r := []string{
				s.PermissionID,
				s.Label,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Role:
		a, err := d.ListRoles()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			r := []string{
				row.RoleID.String(),
				row.Label,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Route:
		a, err := d.ListRoutes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringRoute(row)
			r := []string{
				s.RouteID,
				s.Slug,
				s.Title,
				s.Status,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Session:
		a, err := d.ListSessions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringSession(row)
			r := []string{
				s.SessionID,
				s.UserID,
				s.CreatedAt,
				s.ExpiresAt,
				s.LastAccess,
				s.IpAddress,
				s.UserAgent,
				s.SessionData,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Table:
		a, err := d.ListTables()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringTable(row)
			r := []string{
				s.ID,
				s.Label,
				s.AuthorID,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Token:
		a, err := d.ListTokens()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringToken(row)
			r := []string{
				s.ID,
				s.UserID,
				s.TokenType,
				s.Token,
				s.IssuedAt,
				s.ExpiresAt,
				s.Revoked,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case User:
		a, err := d.ListUsers()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringUser(row)
			r := []string{
				s.UserID,
				s.Username,
				s.Name,
				s.Email,
				s.Hash,
				s.Role,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case User_oauth:
		a, err := d.ListUserOauths()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringUserOauth(row)
			r := []string{
				s.UserOauthID,
				s.UserID,
				s.OauthProvider,
				s.OauthProviderUserID,
				s.AccessToken,
				s.RefreshToken,
				s.TokenExpiresAt,
				s.DateCreated,
			}
			collection = append(collection, r)
		}
		return collection, nil
	}
	return nil, nil
}
