package remote

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// AdminRoute: SDK <-> db
// ---------------------------------------------------------------------------

// adminRouteToDb converts a SDK AdminRoute to a db AdminRoutes.
func adminRouteToDb(s *modula.AdminRoute) db.AdminRoutes {
	return db.AdminRoutes{
		AdminRouteID: types.AdminRouteID(string(s.AdminRouteID)),
		Slug:         types.Slug(string(s.Slug)),
		Title:        s.Title,
		Status:       s.Status,
		AuthorID:     nullUserID(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// adminRouteFromDb converts a db AdminRoutes to a SDK AdminRoute.
func adminRouteFromDb(d db.AdminRoutes) modula.AdminRoute {
	return modula.AdminRoute{
		AdminRouteID: modula.AdminRouteID(string(d.AdminRouteID)),
		Slug:         modula.Slug(string(d.Slug)),
		Title:        d.Title,
		Status:       d.Status,
		AuthorID:     userIDPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// adminRouteCreateFromDb converts db CreateAdminRouteParams to SDK CreateAdminRouteParams.
func adminRouteCreateFromDb(d db.CreateAdminRouteParams) modula.CreateAdminRouteParams {
	return modula.CreateAdminRouteParams{
		Slug:     modula.Slug(string(d.Slug)),
		Title:    d.Title,
		Status:   d.Status,
		AuthorID: userIDPtr(d.AuthorID),
	}
}

// adminRouteUpdateFromDb converts db UpdateAdminRouteParams to SDK UpdateAdminRouteParams.
// Note: db.UpdateAdminRouteParams uses Slug_2 as the old slug identifier;
// the SDK uses AdminRouteID. The caller must set AdminRouteID separately if needed.
func adminRouteUpdateFromDb(d db.UpdateAdminRouteParams) modula.UpdateAdminRouteParams {
	return modula.UpdateAdminRouteParams{
		Slug:     modula.Slug(string(d.Slug)),
		Title:    d.Title,
		Status:   d.Status,
		AuthorID: userIDPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// AdminDatatype: SDK <-> db
// ---------------------------------------------------------------------------

// adminDatatypeToDb converts a SDK AdminDatatype to a db AdminDatatypes.
func adminDatatypeToDb(s *modula.AdminDatatype) db.AdminDatatypes {
	return db.AdminDatatypes{
		AdminDatatypeID: types.AdminDatatypeID(string(s.AdminDatatypeID)),
		ParentID:        nullAdminDatatypeID(s.ParentID),
		Name:            s.Name,
		Label:           s.Label,
		Type:            s.Type,
		AuthorID:        userIDPtrToDb(s.AuthorID),
		DateCreated:     sdkTimestampToDb(s.DateCreated),
		DateModified:    sdkTimestampToDb(s.DateModified),
	}
}

// adminDatatypeFromDb converts a db AdminDatatypes to a SDK AdminDatatype.
func adminDatatypeFromDb(d db.AdminDatatypes) modula.AdminDatatype {
	return modula.AdminDatatype{
		AdminDatatypeID: modula.AdminDatatypeID(string(d.AdminDatatypeID)),
		ParentID:        adminDatatypeIDPtr(d.ParentID),
		Name:            d.Name,
		Label:           d.Label,
		Type:            d.Type,
		AuthorID:        userIDToSdkPtr(d.AuthorID),
		DateCreated:     dbTimestampToSdk(d.DateCreated),
		DateModified:    dbTimestampToSdk(d.DateModified),
	}
}

// adminDatatypeCreateFromDb converts db CreateAdminDatatypeParams to SDK CreateAdminDatatypeParams.
func adminDatatypeCreateFromDb(d db.CreateAdminDatatypeParams) modula.CreateAdminDatatypeParams {
	return modula.CreateAdminDatatypeParams{
		ParentID: adminDatatypeIDPtr(d.ParentID),
		Name:     d.Name,
		Label:    d.Label,
		Type:     d.Type,
		AuthorID: userIDToSdkPtr(d.AuthorID),
	}
}

// adminDatatypeUpdateFromDb converts db UpdateAdminDatatypeParams to SDK UpdateAdminDatatypeParams.
func adminDatatypeUpdateFromDb(d db.UpdateAdminDatatypeParams) modula.UpdateAdminDatatypeParams {
	return modula.UpdateAdminDatatypeParams{
		AdminDatatypeID: modula.AdminDatatypeID(string(d.AdminDatatypeID)),
		ParentID:        adminDatatypeIDPtr(d.ParentID),
		Name:            d.Name,
		Label:           d.Label,
		Type:            d.Type,
		AuthorID:        userIDToSdkPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// AdminField: SDK <-> db
// ---------------------------------------------------------------------------

// adminFieldToDb converts a SDK AdminField to a db AdminFields.
func adminFieldToDb(s *modula.AdminField) db.AdminFields {
	return db.AdminFields{
		AdminFieldID: types.AdminFieldID(string(s.AdminFieldID)),
		ParentID:     nullAdminDatatypeID(s.ParentID),
		SortOrder:    s.SortOrder,
		Name:         s.Name,
		Label:        s.Label,
		Data:         s.Data,
		Validation:   s.Validation,
		UIConfig:     s.UIConfig,
		Type:         types.FieldType(string(s.Type)),
		Translatable: s.Translatable,
		Roles:        rolesToNullableString(s.Roles),
		AuthorID:     nullUserID(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// adminFieldFromDb converts a db AdminFields to a SDK AdminField.
func adminFieldFromDb(d db.AdminFields) modula.AdminField {
	return modula.AdminField{
		AdminFieldID: modula.AdminFieldID(string(d.AdminFieldID)),
		ParentID:     adminDatatypeIDPtr(d.ParentID),
		SortOrder:    d.SortOrder,
		Name:         d.Name,
		Label:        d.Label,
		Data:         d.Data,
		Validation:   d.Validation,
		UIConfig:     d.UIConfig,
		Type:         modula.FieldType(string(d.Type)),
		Translatable: d.Translatable,
		Roles:        nullableStringToRoles(d.Roles),
		AuthorID:     userIDPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// adminFieldCreateFromDb converts db CreateAdminFieldParams to SDK CreateAdminFieldParams.
func adminFieldCreateFromDb(d db.CreateAdminFieldParams) modula.CreateAdminFieldParams {
	return modula.CreateAdminFieldParams{
		ParentID:   adminDatatypeIDPtr(d.ParentID),
		SortOrder:  d.SortOrder,
		Name:       d.Name,
		Label:      d.Label,
		Data:       d.Data,
		Validation: d.Validation,
		UIConfig:   d.UIConfig,
		Type:       modula.FieldType(string(d.Type)),
		Roles:      nullableStringToRoles(d.Roles),
		AuthorID:   userIDPtr(d.AuthorID),
	}
}

// adminFieldUpdateFromDb converts db UpdateAdminFieldParams to SDK UpdateAdminFieldParams.
func adminFieldUpdateFromDb(d db.UpdateAdminFieldParams) modula.UpdateAdminFieldParams {
	return modula.UpdateAdminFieldParams{
		AdminFieldID: modula.AdminFieldID(string(d.AdminFieldID)),
		ParentID:     adminDatatypeIDPtr(d.ParentID),
		SortOrder:    d.SortOrder,
		Name:         d.Name,
		Label:        d.Label,
		Data:         d.Data,
		Validation:   d.Validation,
		UIConfig:     d.UIConfig,
		Type:         modula.FieldType(string(d.Type)),
		Roles:        nullableStringToRoles(d.Roles),
		AuthorID:     userIDPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// AdminFieldType: SDK <-> db
// ---------------------------------------------------------------------------

// adminFieldTypeToDb converts a SDK AdminFieldTypeInfo to a db AdminFieldTypes.
func adminFieldTypeToDb(s *modula.AdminFieldTypeInfo) db.AdminFieldTypes {
	return db.AdminFieldTypes{
		AdminFieldTypeID: types.AdminFieldTypeID(string(s.AdminFieldTypeID)),
		Type:             s.Type,
		Label:            s.Label,
	}
}

// adminFieldTypeFromDb converts a db AdminFieldTypes to a SDK AdminFieldTypeInfo.
func adminFieldTypeFromDb(d db.AdminFieldTypes) modula.AdminFieldTypeInfo {
	return modula.AdminFieldTypeInfo{
		AdminFieldTypeID: modula.AdminFieldTypeID(string(d.AdminFieldTypeID)),
		Type:             d.Type,
		Label:            d.Label,
	}
}

// adminFieldTypeCreateFromDb converts db CreateAdminFieldTypeParams to SDK CreateAdminFieldTypeParams.
func adminFieldTypeCreateFromDb(d db.CreateAdminFieldTypeParams) modula.CreateAdminFieldTypeParams {
	return modula.CreateAdminFieldTypeParams{
		Type:  d.Type,
		Label: d.Label,
	}
}

// adminFieldTypeUpdateFromDb converts db UpdateAdminFieldTypeParams to SDK UpdateAdminFieldTypeParams.
func adminFieldTypeUpdateFromDb(d db.UpdateAdminFieldTypeParams) modula.UpdateAdminFieldTypeParams {
	return modula.UpdateAdminFieldTypeParams{
		AdminFieldTypeID: modula.AdminFieldTypeID(string(d.AdminFieldTypeID)),
		Type:             d.Type,
		Label:            d.Label,
	}
}

// ---------------------------------------------------------------------------
// AdminContentData: SDK <-> db
// ---------------------------------------------------------------------------

// adminContentDataToDb converts a SDK AdminContentData to a db AdminContentData.
func adminContentDataToDb(s *modula.AdminContentData) db.AdminContentData {
	return db.AdminContentData{
		AdminContentDataID: types.AdminContentID(string(s.AdminContentDataID)),
		ParentID:           nullAdminContentID(s.ParentID),
		FirstChildID:       nullAdminContentIDFromString(s.FirstChildID),
		NextSiblingID:      nullAdminContentIDFromString(s.NextSiblingID),
		PrevSiblingID:      nullAdminContentIDFromString(s.PrevSiblingID),
		AdminRouteID:       nullAdminRouteID(s.AdminRouteID),
		AdminDatatypeID:    nullAdminDatatypeID(s.AdminDatatypeID),
		AuthorID:           userIDPtrToDb(s.AuthorID),
		Status:             types.ContentStatus(string(s.Status)),
		PublishedAt:        sdkTimestampPtrToDb(s.PublishedAt),
		PublishedBy:        nullUserID(s.PublishedBy),
		PublishAt:          sdkTimestampPtrToDb(s.PublishAt),
		Revision:           s.Revision,
		DateCreated:        sdkTimestampToDb(s.DateCreated),
		DateModified:       sdkTimestampToDb(s.DateModified),
	}
}

// adminContentDataFromDb converts a db AdminContentData to a SDK AdminContentData.
func adminContentDataFromDb(d db.AdminContentData) modula.AdminContentData {
	var parentID *modula.AdminContentID
	if d.ParentID.Valid {
		id := modula.AdminContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.AdminContentData{
		AdminContentDataID: modula.AdminContentID(string(d.AdminContentDataID)),
		ParentID:           parentID,
		FirstChildID:       nullableAdminContentIDToString(d.FirstChildID),
		NextSiblingID:      nullableAdminContentIDToString(d.NextSiblingID),
		PrevSiblingID:      nullableAdminContentIDToString(d.PrevSiblingID),
		AdminRouteID:       adminRouteIDPtr(d.AdminRouteID),
		AdminDatatypeID:    adminDatatypeIDPtr(d.AdminDatatypeID),
		AuthorID:           userIDToSdkPtr(d.AuthorID),
		Status:             modula.ContentStatus(string(d.Status)),
		PublishedAt:        dbTimestampToSdkPtr(d.PublishedAt),
		PublishedBy:        userIDPtr(d.PublishedBy),
		PublishAt:          dbTimestampToSdkPtr(d.PublishAt),
		Revision:           d.Revision,
		DateCreated:        dbTimestampToSdk(d.DateCreated),
		DateModified:       dbTimestampToSdk(d.DateModified),
	}
}

// adminContentDataCreateFromDb converts db CreateAdminContentDataParams to SDK CreateAdminContentDataParams.
func adminContentDataCreateFromDb(d db.CreateAdminContentDataParams) modula.CreateAdminContentDataParams {
	var parentID *modula.AdminContentID
	if d.ParentID.Valid {
		id := modula.AdminContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.CreateAdminContentDataParams{
		ParentID:        parentID,
		FirstChildID:    nullableAdminContentIDToString(d.FirstChildID),
		NextSiblingID:   nullableAdminContentIDToString(d.NextSiblingID),
		PrevSiblingID:   nullableAdminContentIDToString(d.PrevSiblingID),
		AdminRouteID:    adminRouteIDPtr(d.AdminRouteID),
		AdminDatatypeID: adminDatatypeIDPtr(d.AdminDatatypeID),
		AuthorID:        userIDToSdkPtr(d.AuthorID),
		Status:          modula.ContentStatus(string(d.Status)),
	}
}

// adminContentDataUpdateFromDb converts db UpdateAdminContentDataParams to SDK UpdateAdminContentDataParams.
func adminContentDataUpdateFromDb(d db.UpdateAdminContentDataParams) modula.UpdateAdminContentDataParams {
	var parentID *modula.AdminContentID
	if d.ParentID.Valid {
		id := modula.AdminContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.UpdateAdminContentDataParams{
		AdminContentDataID: modula.AdminContentID(string(d.AdminContentDataID)),
		ParentID:           parentID,
		FirstChildID:       nullableAdminContentIDToString(d.FirstChildID),
		NextSiblingID:      nullableAdminContentIDToString(d.NextSiblingID),
		PrevSiblingID:      nullableAdminContentIDToString(d.PrevSiblingID),
		AdminRouteID:       adminRouteIDPtr(d.AdminRouteID),
		AdminDatatypeID:    adminDatatypeIDPtr(d.AdminDatatypeID),
		AuthorID:           userIDToSdkPtr(d.AuthorID),
		Status:             modula.ContentStatus(string(d.Status)),
	}
}

// ---------------------------------------------------------------------------
// AdminContentField: SDK <-> db
// ---------------------------------------------------------------------------

// adminContentFieldToDb converts a SDK AdminContentField to a db AdminContentFields.
func adminContentFieldToDb(s *modula.AdminContentField) db.AdminContentFields {
	return db.AdminContentFields{
		AdminContentFieldID: types.AdminContentFieldID(string(s.AdminContentFieldID)),
		AdminRouteID:        nullAdminRouteID(s.AdminRouteID),
		AdminContentDataID:  nullAdminContentID(s.AdminContentDataID),
		AdminFieldID:        nullAdminFieldID(s.AdminFieldID),
		AdminFieldValue:     s.AdminFieldValue,
		Locale:              s.Locale,
		AuthorID:            userIDPtrToDb(s.AuthorID),
		DateCreated:         sdkTimestampToDb(s.DateCreated),
		DateModified:        sdkTimestampToDb(s.DateModified),
	}
}

// adminContentFieldFromDb converts a db AdminContentFields to a SDK AdminContentField.
func adminContentFieldFromDb(d db.AdminContentFields) modula.AdminContentField {
	return modula.AdminContentField{
		AdminContentFieldID: modula.AdminContentFieldID(string(d.AdminContentFieldID)),
		AdminRouteID:        adminRouteIDPtr(d.AdminRouteID),
		AdminContentDataID:  adminContentIDPtr(d.AdminContentDataID),
		AdminFieldID:        adminFieldIDPtr(d.AdminFieldID),
		AdminFieldValue:     d.AdminFieldValue,
		Locale:              d.Locale,
		AuthorID:            userIDToSdkPtr(d.AuthorID),
		DateCreated:         dbTimestampToSdk(d.DateCreated),
		DateModified:        dbTimestampToSdk(d.DateModified),
	}
}
