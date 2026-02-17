package cli

import (
	"database/sql"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// Parse scans sql.Rows into appropriate structs based on the DBTable type.
// Returns a slice of results or falls back to map[string]any for unknown tables.
func Parse(rows *sql.Rows, table db.DBTable) (any, error) {
	if rows == nil {
		return nil, fmt.Errorf("rows cannot be nil")
	}
	defer utility.HandleRowsCloseDeferErr(rows)

	switch table {
	case db.User:
		return parseUsers(rows)
	case db.Role:
		return parseRoles(rows)
	case db.Permission:
		return parsePermissions(rows)
	case db.Session:
		return parseSessions(rows)
	case db.Token:
		return parseTokens(rows)
	case db.User_oauth:
		return parseUserOauth(rows)
	case db.Route:
		return parseRoutes(rows)
	case db.Admin_route:
		return parseAdminRoutes(rows)
	case db.Field:
		return parseFields(rows)
	case db.Admin_field:
		return parseAdminFields(rows)
	case db.Datatype:
		return parseDatatypes(rows)
	case db.Admin_datatype:
		return parseAdminDatatypes(rows)
	case db.Datatype_fields:
		return parseDatatypeFields(rows)
	case db.Admin_datatype_fields:
		return parseAdminDatatypeFields(rows)
	case db.Content_data:
		return parseContentData(rows)
	case db.Admin_content_data:
		return parseAdminContentData(rows)
	case db.Content_fields:
		return parseContentFields(rows)
	case db.Admin_content_fields:
		return parseAdminContentFields(rows)
	case db.MediaT:
		return parseMedia(rows)
	case db.Media_dimension:
		return parseMediaDimensions(rows)
	case db.Table:
		return parseTables(rows)
	default:
		return parseGeneric(rows)
	}
}

// parseUsers scans rows into Users structs
func parseUsers(rows *sql.Rows) ([]db.Users, error) {
	var results []db.Users

	for rows.Next() {
		var user db.Users
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.Name,
			&user.Email,
			&user.Hash,
			&user.Role,
			&user.DateCreated,
			&user.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		results = append(results, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseRoles scans rows into Roles structs
func parseRoles(rows *sql.Rows) ([]db.Roles, error) {
	var results []db.Roles

	for rows.Next() {
		var role db.Roles
		err := rows.Scan(
			&role.RoleID,
			&role.Label,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %v", err)
		}
		results = append(results, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parsePermissions scans rows into Permissions structs
func parsePermissions(rows *sql.Rows) ([]db.Permissions, error) {
	var results []db.Permissions

	for rows.Next() {
		var permission db.Permissions
		err := rows.Scan(
			&permission.PermissionID,
			&permission.Label,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %v", err)
		}
		results = append(results, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseSessions scans rows into Sessions structs
func parseSessions(rows *sql.Rows) ([]db.Sessions, error) {
	var results []db.Sessions

	for rows.Next() {
		var session db.Sessions
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.LastAccess,
			&session.IpAddress,
			&session.UserAgent,
			&session.SessionData,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %v", err)
		}
		results = append(results, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseTokens scans rows into Tokens structs
func parseTokens(rows *sql.Rows) ([]db.Tokens, error) {
	var results []db.Tokens

	for rows.Next() {
		var token db.Tokens
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenType,
			&token.Token,
			&token.IssuedAt,
			&token.ExpiresAt,
			&token.Revoked,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token: %v", err)
		}
		results = append(results, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseUserOauth scans rows into UserOauth structs
func parseUserOauth(rows *sql.Rows) ([]db.UserOauth, error) {
	var results []db.UserOauth

	for rows.Next() {
		var oauth db.UserOauth
		err := rows.Scan(
			&oauth.UserOauthID,
			&oauth.UserID,
			&oauth.OauthProvider,
			&oauth.OauthProviderUserID,
			&oauth.AccessToken,
			&oauth.RefreshToken,
			&oauth.TokenExpiresAt,
			&oauth.DateCreated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user oauth: %v", err)
		}
		results = append(results, oauth)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseRoutes scans rows into Routes structs
func parseRoutes(rows *sql.Rows) ([]db.Routes, error) {
	var results []db.Routes

	for rows.Next() {
		var route db.Routes
		err := rows.Scan(
			&route.RouteID,
			&route.Slug,
			&route.Title,
			&route.Status,
			&route.AuthorID,
			&route.DateCreated,
			&route.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route: %v", err)
		}
		results = append(results, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminRoutes scans rows into AdminRoutes structs
func parseAdminRoutes(rows *sql.Rows) ([]db.AdminRoutes, error) {
	var results []db.AdminRoutes

	for rows.Next() {
		var route db.AdminRoutes
		err := rows.Scan(
			&route.AdminRouteID,
			&route.Slug,
			&route.Title,
			&route.Status,
			&route.AuthorID,
			&route.DateCreated,
			&route.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin route: %v", err)
		}
		results = append(results, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseFields scans rows into Fields structs
func parseFields(rows *sql.Rows) ([]db.Fields, error) {
	var results []db.Fields

	for rows.Next() {
		var field db.Fields
		err := rows.Scan(
			&field.FieldID,
			&field.ParentID,
			&field.Label,
			&field.Data,
			&field.Type,
			&field.AuthorID,
			&field.DateCreated,
			&field.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field: %v", err)
		}
		results = append(results, field)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminFields scans rows into AdminFields structs
func parseAdminFields(rows *sql.Rows) ([]db.AdminFields, error) {
	var results []db.AdminFields

	for rows.Next() {
		var field db.AdminFields
		err := rows.Scan(
			&field.AdminFieldID,
			&field.ParentID,
			&field.Label,
			&field.Data,
			&field.Type,
			&field.AuthorID,
			&field.DateCreated,
			&field.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin field: %v", err)
		}
		results = append(results, field)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseDatatypes scans rows into Datatypes structs
func parseDatatypes(rows *sql.Rows) ([]db.Datatypes, error) {
	var results []db.Datatypes

	for rows.Next() {
		var datatype db.Datatypes
		err := rows.Scan(
			&datatype.DatatypeID,
			&datatype.ParentID,
			&datatype.Label,
			&datatype.Type,
			&datatype.AuthorID,
			&datatype.DateCreated,
			&datatype.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan datatype: %v", err)
		}
		results = append(results, datatype)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminDatatypes scans rows into AdminDatatypes structs
func parseAdminDatatypes(rows *sql.Rows) ([]db.AdminDatatypes, error) {
	var results []db.AdminDatatypes

	for rows.Next() {
		var datatype db.AdminDatatypes
		err := rows.Scan(
			&datatype.AdminDatatypeID,
			&datatype.ParentID,
			&datatype.Label,
			&datatype.Type,
			&datatype.AuthorID,
			&datatype.DateCreated,
			&datatype.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin datatype: %v", err)
		}
		results = append(results, datatype)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseDatatypeFields scans rows into DatatypeFields structs
func parseDatatypeFields(rows *sql.Rows) ([]db.DatatypeFields, error) {
	var results []db.DatatypeFields

	for rows.Next() {
		var dtField db.DatatypeFields
		err := rows.Scan(
			&dtField.ID,
			&dtField.DatatypeID,
			&dtField.FieldID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan datatype field: %v", err)
		}
		results = append(results, dtField)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminDatatypeFields scans rows into AdminDatatypeFields structs
func parseAdminDatatypeFields(rows *sql.Rows) ([]db.AdminDatatypeFields, error) {
	var results []db.AdminDatatypeFields

	for rows.Next() {
		var dtField db.AdminDatatypeFields
		err := rows.Scan(
			&dtField.ID,
			&dtField.AdminDatatypeID,
			&dtField.AdminFieldID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin datatype field: %v", err)
		}
		results = append(results, dtField)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseContentData scans rows into ContentData structs
func parseContentData(rows *sql.Rows) ([]db.ContentData, error) {
	var results []db.ContentData

	for rows.Next() {
		var content db.ContentData
		err := rows.Scan(
			&content.ContentDataID,
			&content.ParentID,
			&content.RouteID,
			&content.DatatypeID,
			&content.AuthorID,
			&content.DateCreated,
			&content.DateModified,
			&content.FirstChildID,
			&content.NextSiblingID,
			&content.PrevSiblingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content data: %v", err)
		}
		results = append(results, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminContentData scans rows into AdminContentData structs
func parseAdminContentData(rows *sql.Rows) ([]db.AdminContentData, error) {
	var results []db.AdminContentData

	for rows.Next() {
		var content db.AdminContentData
		err := rows.Scan(
			&content.AdminContentDataID,
			&content.ParentID,
			&content.AdminRouteID,
			&content.AdminDatatypeID,
			&content.AuthorID,
			&content.DateCreated,
			&content.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin content data: %v", err)
		}
		results = append(results, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseContentFields scans rows into ContentFields structs
func parseContentFields(rows *sql.Rows) ([]db.ContentFields, error) {
	var results []db.ContentFields

	for rows.Next() {
		var contentField db.ContentFields
		err := rows.Scan(
			&contentField.ContentFieldID,
			&contentField.RouteID,
			&contentField.ContentDataID,
			&contentField.FieldID,
			&contentField.FieldValue,
			&contentField.AuthorID,
			&contentField.DateCreated,
			&contentField.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content field: %v", err)
		}
		results = append(results, contentField)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseAdminContentFields scans rows into AdminContentFields structs
func parseAdminContentFields(rows *sql.Rows) ([]db.AdminContentFields, error) {
	var results []db.AdminContentFields

	for rows.Next() {
		var contentField db.AdminContentFields
		err := rows.Scan(
			&contentField.AdminContentFieldID,
			&contentField.AdminRouteID,
			&contentField.AdminContentDataID,
			&contentField.AdminFieldID,
			&contentField.AdminFieldValue,
			&contentField.AuthorID,
			&contentField.DateCreated,
			&contentField.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin content field: %v", err)
		}
		results = append(results, contentField)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseMedia scans rows into Media structs
func parseMedia(rows *sql.Rows) ([]db.Media, error) {
	var results []db.Media

	for rows.Next() {
		var media db.Media
		err := rows.Scan(
			&media.MediaID,
			&media.Name,
			&media.DisplayName,
			&media.Alt,
			&media.Caption,
			&media.Description,
			&media.Class,
			&media.Mimetype,
			&media.Dimensions,
			&media.URL,
			&media.Srcset,
			&media.AuthorID,
			&media.DateCreated,
			&media.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media: %v", err)
		}
		results = append(results, media)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseMediaDimensions scans rows into MediaDimensions structs
func parseMediaDimensions(rows *sql.Rows) ([]db.MediaDimensions, error) {
	var results []db.MediaDimensions

	for rows.Next() {
		var dimension db.MediaDimensions
		err := rows.Scan(
			&dimension.MdID,
			&dimension.Label,
			&dimension.Width,
			&dimension.Height,
			&dimension.AspectRatio,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media dimension: %v", err)
		}
		results = append(results, dimension)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseTables scans rows into Tables structs
func parseTables(rows *sql.Rows) ([]db.Tables, error) {
	var results []db.Tables

	for rows.Next() {
		var table db.Tables
		err := rows.Scan(
			&table.ID,
			&table.Label,
			&table.AuthorID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table: %v", err)
		}
		results = append(results, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// parseGeneric scans rows into []map[string]any for unknown table types
func parseGeneric(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if val == nil {
				row[col] = nil
			} else {
				switch v := val.(type) {
				case []byte:
					row[col] = string(v)
				default:
					row[col] = v
				}
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}
