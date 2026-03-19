package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/types"
)

// AssembleUserFullView fetches a user and all related entities,
// returning a composed view. Uses bounded sequential queries:
// 1 user + 1 role + 1 oauth + 1 ssh keys + 1 session + 1 tokens.
func AssembleUserFullView(d DbDriver, userID types.UserID) (*UserFullView, error) {
	user, err := d.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	view := UserFullView{
		UserID:       user.UserID,
		Username:     user.Username,
		Name:         user.Name,
		Email:        user.Email,
		RoleID:       user.Role,
		DateCreated:  user.DateCreated,
		DateModified: user.DateModified,
		SshKeys:      []UserSshKeyView{},
		Tokens:       []TokenView{},
	}

	if user.Role != "" {
		role, roleErr := d.GetRole(types.RoleID(user.Role))
		if roleErr == nil {
			view.RoleLabel = role.Label
		}
	}

	nullUserID := types.NullableUserID{ID: userID, Valid: true}

	oauth, err := d.GetUserOauthByUserId(nullUserID)
	if err == nil {
		ov := MapUserOauthView(*oauth)
		view.Oauth = &ov
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get user oauth: %w", err)
	}

	sshKeys, err := d.ListUserSshKeys(nullUserID)
	if err != nil {
		return nil, fmt.Errorf("list ssh keys: %w", err)
	}
	for _, k := range *sshKeys {
		view.SshKeys = append(view.SshKeys, MapUserSshKeyView(k))
	}

	session, err := d.GetSessionByUserId(nullUserID)
	if err == nil {
		sv := MapSessionView(*session)
		view.Sessions = &sv
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get session: %w", err)
	}

	tokens, err := d.GetTokenByUserId(nullUserID)
	if err != nil {
		return nil, fmt.Errorf("get tokens: %w", err)
	}
	for _, t := range *tokens {
		view.Tokens = append(view.Tokens, MapTokenView(t))
	}

	return &view, nil
}

// AssembleDatatypeFullView fetches a datatype and all its field definitions,
// returning a composed view. Uses bounded sequential queries:
// 1 datatype + 0-1 author + 1 fields JOIN.
func AssembleDatatypeFullView(d DbDriver, datatypeID types.DatatypeID) (*DatatypeFullView, error) {
	dt, err := d.GetDatatype(datatypeID)
	if err != nil {
		return nil, fmt.Errorf("get datatype: %w", err)
	}

	view := DatatypeFullView{
		DatatypeID:   dt.DatatypeID,
		Label:        dt.Label,
		Type:         dt.Type,
		ParentID:     dt.ParentID,
		DateCreated:  dt.DateCreated,
		DateModified: dt.DateModified,
		Fields:       []DatatypeFieldView{},
	}

	user, err := d.GetUser(dt.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("get author %s: %w", dt.AuthorID, err)
	}
	av := MapAuthorView(*user)
	view.Author = &av

	rows, err := d.ListFieldsWithSortOrderByDatatypeID(types.NullableDatatypeID{ID: datatypeID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list fields with sort order: %w", err)
	}

	for _, row := range *rows {
		view.Fields = append(view.Fields, MapDatatypeFieldView(row))
	}

	return &view, nil
}

// AssembleContentDataView fetches a content data item and its relations,
// returning a composed view. Uses bounded sequential queries:
// 1 content + 1 author + 1 datatype + 1 list fields with field defs (JOIN).
func AssembleContentDataView(d DbDriver, contentID types.ContentID) (*ContentDataView, error) {
	cd, err := d.GetContentData(contentID)
	if err != nil {
		return nil, fmt.Errorf("get content data: %w", err)
	}

	view := ContentDataView{
		ContentDataID: cd.ContentDataID,
		Status:        cd.Status,
		DateCreated:   cd.DateCreated,
		DateModified:  cd.DateModified,
		Fields:        []FieldView{},
	}

	user, err := d.GetUser(cd.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("get author %s: %w", cd.AuthorID, err)
	}
	av := MapAuthorView(*user)
	view.Author = &av

	if cd.DatatypeID.Valid {
		dt, err := d.GetDatatype(cd.DatatypeID.ID)
		if err != nil {
			return nil, fmt.Errorf("get datatype %s: %w", cd.DatatypeID.ID, err)
		}
		dv := MapDatatypeView(*dt)
		view.Datatype = &dv
	}

	contentDataID := types.NullableContentID{ID: cd.ContentDataID, Valid: true}
	rows, err := d.ListContentFieldsWithFieldByContentData(contentDataID)
	if err != nil {
		return nil, fmt.Errorf("list content fields with field defs: %w", err)
	}

	view.Fields = AssembleFieldViews(*rows)

	return &view, nil
}

// AssembleFieldViews converts JOIN rows into a FieldView slice without N+1.
func AssembleFieldViews(rows []ContentFieldWithFieldRow) []FieldView {
	result := make([]FieldView, 0, len(rows))
	for _, row := range rows {
		result = append(result, MapFieldViewFromRow(row))
	}
	return result
}

// AssembleMediaFullView converts a Media entity into a MediaFullView with author.
func AssembleMediaFullView(d DbDriver, m Media) MediaFullView {
	view := MapMediaFullView(m)
	if m.AuthorID.Valid {
		user, err := d.GetUser(m.AuthorID.ID)
		if err == nil {
			av := MapAuthorView(*user)
			view.Author = &av
		}
	}
	return view
}

// AssembleMediaFullListView converts a slice of Media into MediaFullViews
// with authors resolved via a single-pass cache.
func AssembleMediaFullListView(d DbDriver, items []Media) []MediaFullView {
	authorCache := make(map[types.UserID]*AuthorView)
	views := make([]MediaFullView, 0, len(items))
	for _, m := range items {
		view := MapMediaFullView(m)
		if m.AuthorID.Valid {
			av, ok := authorCache[m.AuthorID.ID]
			if !ok {
				user, err := d.GetUser(m.AuthorID.ID)
				if err == nil {
					mapped := MapAuthorView(*user)
					av = &mapped
					authorCache[m.AuthorID.ID] = av
				}
			}
			if av != nil {
				copied := *av
				view.Author = &copied
			}
		}
		views = append(views, view)
	}
	return views
}

// AssembleRouteFullView fetches a route and its content tree, returning a composed view.
// Uses bounded sequential queries: 1 route + 0-1 author + 1 tree query.
func AssembleRouteFullView(d DbDriver, routeID types.RouteID) (*RouteFullView, error) {
	route, err := d.GetRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("get route: %w", err)
	}

	view := RouteFullView{
		RouteID:      route.RouteID,
		Slug:         route.Slug,
		Title:        route.Title,
		Status:       route.Status,
		ContentTree:  []RouteContentNodeView{},
		DateCreated:  route.DateCreated,
		DateModified: route.DateModified,
	}

	if route.AuthorID.Valid {
		user, userErr := d.GetUser(route.AuthorID.ID)
		if userErr == nil {
			av := MapAuthorView(*user)
			view.Author = &av
		}
	}

	nullRouteID := types.NullableRouteID{ID: routeID, Valid: true}
	rows, err := d.GetContentTreeByRoute(nullRouteID)
	if err != nil {
		return nil, fmt.Errorf("get content tree: %w", err)
	}

	for _, row := range *rows {
		view.ContentTree = append(view.ContentTree, RouteContentNodeView{
			ContentDataID: row.ContentDataID,
			ParentID:      row.ParentID,
			DatatypeLabel: row.DatatypeLabel,
			DatatypeType:  row.DatatypeType,
			Status:        row.Status,
			DateCreated:   row.DateCreated,
			DateModified:  row.DateModified,
		})
	}

	return &view, nil
}

// AssembleRecentActivity converts change events into activity views with actor info.
// Uses a single-pass author cache to avoid N+1 user lookups.
func AssembleRecentActivity(d DbDriver, events []ChangeEvent) []ActivityEventView {
	authorCache := make(map[types.UserID]*AuthorView)
	views := make([]ActivityEventView, 0, len(events))
	for _, ev := range events {
		view := ActivityEventView{
			EventID:       ev.EventID,
			TableName:     ev.TableName,
			RecordID:      ev.RecordID,
			Operation:     ev.Operation,
			Action:        ev.Action,
			WallTimestamp: ev.WallTimestamp,
		}
		if ev.UserID.Valid {
			av, ok := authorCache[ev.UserID.ID]
			if !ok {
				user, err := d.GetUser(ev.UserID.ID)
				if err == nil {
					mapped := MapAuthorView(*user)
					av = &mapped
					authorCache[ev.UserID.ID] = av
				}
			}
			if av != nil {
				copied := *av
				view.Actor = &copied
			}
		}
		views = append(views, view)
	}
	return views
}

// AssembleDatatypeListView builds a list of datatypes with field counts and parent labels.
// Uses bounded queries: 1 list datatypes + N field-count queries (cached by datatype).
func AssembleDatatypeListView(d DbDriver, datatypes []Datatypes) []DatatypeListItemView {
	// Build parent label lookup
	parentLabels := make(map[types.DatatypeID]string, len(datatypes))
	for _, dt := range datatypes {
		parentLabels[dt.DatatypeID] = dt.Label
	}

	authorCache := make(map[types.UserID]*AuthorView)
	views := make([]DatatypeListItemView, 0, len(datatypes))
	for _, dt := range datatypes {
		view := DatatypeListItemView{
			DatatypeID:   dt.DatatypeID,
			Label:        dt.Label,
			Type:         dt.Type,
			ParentID:     dt.ParentID,
			DateCreated:  dt.DateCreated,
			DateModified: dt.DateModified,
		}

		if dt.ParentID.Valid {
			view.ParentLabel = parentLabels[dt.ParentID.ID]
		}

		fields, err := d.ListFieldsWithSortOrderByDatatypeID(types.NullableDatatypeID{ID: dt.DatatypeID, Valid: true})
		if err == nil && fields != nil {
			view.FieldCount = len(*fields)
		}

		av, ok := authorCache[dt.AuthorID]
		if !ok {
			user, err := d.GetUser(dt.AuthorID)
			if err == nil {
				mapped := MapAuthorView(*user)
				av = &mapped
				authorCache[dt.AuthorID] = av
			}
		}
		if av != nil {
			copied := *av
			view.Author = &copied
		}

		views = append(views, view)
	}
	return views
}

// AssembleAdminDatatypeFullView fetches an admin datatype and its field definitions.
// Uses bounded sequential queries: 1 datatype + 0-1 author + 1 fields list.
func AssembleAdminDatatypeFullView(d DbDriver, id types.AdminDatatypeID) (*AdminDatatypeFullView, error) {
	dt, err := d.GetAdminDatatypeById(id)
	if err != nil {
		return nil, fmt.Errorf("get admin datatype: %w", err)
	}

	view := AdminDatatypeFullView{
		AdminDatatypeID: dt.AdminDatatypeID,
		Label:           dt.Label,
		Type:            dt.Type,
		ParentID:        dt.ParentID,
		SortOrder:       dt.SortOrder,
		DateCreated:     dt.DateCreated,
		DateModified:    dt.DateModified,
		Fields:          []AdminFieldView{},
	}

	user, userErr := d.GetUser(dt.AuthorID)
	if userErr == nil {
		av := MapAuthorView(*user)
		view.Author = &av
	}

	fields, err := d.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list admin fields: %w", err)
	}
	for _, f := range *fields {
		view.Fields = append(view.Fields, AdminFieldView{
			AdminFieldID: f.AdminFieldID,
			Label:        f.Label,
			Type:         f.Type,
			Data:         f.Data,
			ValidationID: f.ValidationID,
			UIConfig:     f.UIConfig,
			SortOrder:    f.SortOrder,
		})
	}

	return &view, nil
}

// AssembleAdminContentDataView fetches an admin content data item and its relations.
// Uses bounded sequential queries: 1 content + 1 author + 0-1 datatype + 1 fields JOIN.
func AssembleAdminContentDataView(d DbDriver, contentID types.AdminContentID) (*AdminContentDataView, error) {
	cd, err := d.GetAdminContentData(contentID)
	if err != nil {
		return nil, fmt.Errorf("get admin content data: %w", err)
	}

	view := AdminContentDataView{
		AdminContentDataID: cd.AdminContentDataID,
		Status:             cd.Status,
		Revision:           cd.Revision,
		DateCreated:        cd.DateCreated,
		DateModified:       cd.DateModified,
		Fields:             []AdminContentFieldView{},
	}

	user, err := d.GetUser(cd.AuthorID)
	if err == nil {
		av := MapAuthorView(*user)
		view.Author = &av
	}

	if cd.AdminDatatypeID.Valid {
		dt, dtErr := d.GetAdminDatatypeById(cd.AdminDatatypeID.ID)
		if dtErr == nil {
			view.Datatype = &AdminDatatypeView{
				AdminDatatypeID: dt.AdminDatatypeID,
				Label:           dt.Label,
				Type:            dt.Type,
			}
		}
	}

	nullID := types.NullableAdminContentID{ID: cd.AdminContentDataID, Valid: true}
	rows, err := d.ListAdminContentFieldsWithFieldByContentData(nullID)
	if err != nil {
		return nil, fmt.Errorf("list admin content fields: %w", err)
	}
	for _, row := range *rows {
		view.Fields = append(view.Fields, AdminContentFieldView{
			AdminFieldID: row.FAdminFieldID,
			Label:        row.FLabel,
			Type:         row.FType,
			Value:        row.AdminFieldValue,
		})
	}

	return &view, nil
}

// GroupBy groups a slice by a key function, preserving insertion order of keys.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range items {
		k := key(item)
		result[k] = append(result[k], item)
	}
	return result
}
