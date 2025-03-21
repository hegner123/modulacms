package db

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func GenericList(t DBTable, d DbDriver) (*[][]string, error) {
	f, _ := tea.LogToFile("debug.log", "debug")
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
				s.AdminRouteID,
				s.AdminDatatypeID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
	case Admin_datatype:
		a, err := d.ListAdminDatatypes()
		if err != nil {
			return nil, err
		}
        fmt.Println(f,"help")
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.RouteID,
				s.DatatypeID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
		return &collection, nil
	case Role:
		a, err := d.ListRoles()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringRole(row)
			r := []string{
				s.RoleID,
				s.Label,
				s.Permissions,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
				s.Author,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return &collection, nil
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
		return &collection, nil
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
		return &collection, nil
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
		return &collection, nil
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
		return &collection, nil
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
		return &collection, nil
	}
	return nil, nil
}
