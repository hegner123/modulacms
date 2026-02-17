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

	if dt.AuthorID.Valid {
		user, err := d.GetUser(dt.AuthorID.ID)
		if err != nil {
			return nil, fmt.Errorf("get author %s: %w", dt.AuthorID.ID, err)
		}
		av := MapAuthorView(*user)
		view.Author = &av
	}

	rows, err := d.ListFieldsWithSortOrderByDatatypeID(datatypeID)
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

	if cd.AuthorID.Valid {
		user, err := d.GetUser(cd.AuthorID.ID)
		if err != nil {
			return nil, fmt.Errorf("get author %s: %w", cd.AuthorID.ID, err)
		}
		av := MapAuthorView(*user)
		view.Author = &av
	}

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

// GroupBy groups a slice by a key function, preserving insertion order of keys.
func GroupBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range items {
		k := key(item)
		result[k] = append(result[k], item)
	}
	return result
}
