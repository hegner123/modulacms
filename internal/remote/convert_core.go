package remote

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// Route: SDK <-> db
// ---------------------------------------------------------------------------

// routeToDb converts a SDK Route to a db Routes.
func routeToDb(s *modula.Route) db.Routes {
	return db.Routes{
		RouteID:      types.RouteID(string(s.RouteID)),
		Slug:         types.Slug(string(s.Slug)),
		Title:        s.Title,
		Status:       s.Status,
		AuthorID:     nullUserID(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// routeFromDb converts a db Routes to a SDK Route.
func routeFromDb(d db.Routes) modula.Route {
	return modula.Route{
		RouteID:      modula.RouteID(string(d.RouteID)),
		Slug:         modula.Slug(string(d.Slug)),
		Title:        d.Title,
		Status:       d.Status,
		AuthorID:     userIDPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// routeCreateFromDb converts db CreateRouteParams to SDK CreateRouteParams.
func routeCreateFromDb(d db.CreateRouteParams) modula.CreateRouteParams {
	return modula.CreateRouteParams{
		Slug:     modula.Slug(string(d.Slug)),
		Title:    d.Title,
		Status:   d.Status,
		AuthorID: userIDPtr(d.AuthorID),
	}
}

// routeUpdateFromDb converts db UpdateRouteParams to SDK UpdateRouteParams.
// Note: db.UpdateRouteParams uses Slug_2 as the old slug identifier;
// the SDK uses RouteID. The caller must set RouteID separately if needed.
func routeUpdateFromDb(d db.UpdateRouteParams) modula.UpdateRouteParams {
	return modula.UpdateRouteParams{
		Slug:     modula.Slug(string(d.Slug)),
		Title:    d.Title,
		Status:   d.Status,
		AuthorID: userIDPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// Datatype: SDK <-> db
// ---------------------------------------------------------------------------

// datatypeToDb converts a SDK Datatype to a db Datatypes.
func datatypeToDb(s *modula.Datatype) db.Datatypes {
	return db.Datatypes{
		DatatypeID:   types.DatatypeID(string(s.DatatypeID)),
		ParentID:     nullDatatypeID(s.ParentID),
		Name:         s.Name,
		Label:        s.Label,
		Type:         s.Type,
		AuthorID:     userIDPtrToDb(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// datatypeFromDb converts a db Datatypes to a SDK Datatype.
func datatypeFromDb(d db.Datatypes) modula.Datatype {
	return modula.Datatype{
		DatatypeID:   modula.DatatypeID(string(d.DatatypeID)),
		ParentID:     datatypeIDPtr(d.ParentID),
		Name:         d.Name,
		Label:        d.Label,
		Type:         d.Type,
		AuthorID:     userIDToSdkPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// datatypeCreateFromDb converts db CreateDatatypeParams to SDK CreateDatatypeParams.
func datatypeCreateFromDb(d db.CreateDatatypeParams) modula.CreateDatatypeParams {
	var dtID *modula.DatatypeID
	if !d.DatatypeID.IsZero() {
		id := modula.DatatypeID(string(d.DatatypeID))
		dtID = &id
	}
	return modula.CreateDatatypeParams{
		DatatypeID: dtID,
		ParentID:   datatypeIDPtr(d.ParentID),
		Name:       d.Name,
		Label:      d.Label,
		Type:       d.Type,
		AuthorID:   userIDToSdkPtr(d.AuthorID),
	}
}

// datatypeUpdateFromDb converts db UpdateDatatypeParams to SDK UpdateDatatypeParams.
func datatypeUpdateFromDb(d db.UpdateDatatypeParams) modula.UpdateDatatypeParams {
	return modula.UpdateDatatypeParams{
		DatatypeID: modula.DatatypeID(string(d.DatatypeID)),
		ParentID:   datatypeIDPtr(d.ParentID),
		Name:       d.Name,
		Label:      d.Label,
		Type:       d.Type,
		AuthorID:   userIDToSdkPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// Field: SDK <-> db
// ---------------------------------------------------------------------------

// fieldToDb converts a SDK Field to a db Fields.
func fieldToDb(s *modula.Field) db.Fields {
	return db.Fields{
		FieldID:      types.FieldID(string(s.FieldID)),
		ParentID:     nullDatatypeID(s.ParentID),
		SortOrder:    s.SortOrder,
		Name:         s.Name,
		Label:        s.Label,
		Data:         s.Data,
		ValidationID: types.NullableValidationID{},
		UIConfig:     s.UIConfig,
		Type:         types.FieldType(string(s.Type)),
		Translatable: s.Translatable,
		Roles:        rolesToNullableString(s.Roles),
		AuthorID:     nullUserID(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// fieldFromDb converts a db Fields to a SDK Field.
func fieldFromDb(d db.Fields) modula.Field {
	return modula.Field{
		FieldID:      modula.FieldID(string(d.FieldID)),
		ParentID:     datatypeIDPtr(d.ParentID),
		SortOrder:    d.SortOrder,
		Name:         d.Name,
		Label:        d.Label,
		Data:         d.Data,
		ValidationID: "",
		UIConfig:     d.UIConfig,
		Type:         modula.FieldType(string(d.Type)),
		Translatable: d.Translatable,
		Roles:        nullableStringToRoles(d.Roles),
		AuthorID:     userIDPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// fieldCreateFromDb converts db CreateFieldParams to SDK CreateFieldParams.
func fieldCreateFromDb(d db.CreateFieldParams) modula.CreateFieldParams {
	var fID *modula.FieldID
	if !d.FieldID.IsZero() {
		id := modula.FieldID(string(d.FieldID))
		fID = &id
	}
	return modula.CreateFieldParams{
		FieldID:    fID,
		ParentID:   datatypeIDPtr(d.ParentID),
		SortOrder:  d.SortOrder,
		Name:       d.Name,
		Label:      d.Label,
		Data:       d.Data,
		ValidationID: "",
		UIConfig:   d.UIConfig,
		Type:       modula.FieldType(string(d.Type)),
		Roles:      nullableStringToRoles(d.Roles),
		AuthorID:   userIDPtr(d.AuthorID),
	}
}

// fieldUpdateFromDb converts db UpdateFieldParams to SDK UpdateFieldParams.
func fieldUpdateFromDb(d db.UpdateFieldParams) modula.UpdateFieldParams {
	return modula.UpdateFieldParams{
		FieldID:    modula.FieldID(string(d.FieldID)),
		ParentID:   datatypeIDPtr(d.ParentID),
		SortOrder:  d.SortOrder,
		Name:       d.Name,
		Label:      d.Label,
		Data:       d.Data,
		ValidationID: "",
		UIConfig:   d.UIConfig,
		Type:       modula.FieldType(string(d.Type)),
		Roles:      nullableStringToRoles(d.Roles),
		AuthorID:   userIDPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// FieldType: SDK <-> db
// ---------------------------------------------------------------------------

// fieldTypeToDb converts a SDK FieldTypeInfo to a db FieldTypes.
func fieldTypeToDb(s *modula.FieldTypeInfo) db.FieldTypes {
	return db.FieldTypes{
		FieldTypeID: types.FieldTypeID(string(s.FieldTypeID)),
		Type:        s.Type,
		Label:       s.Label,
	}
}

// fieldTypeFromDb converts a db FieldTypes to a SDK FieldTypeInfo.
func fieldTypeFromDb(d db.FieldTypes) modula.FieldTypeInfo {
	return modula.FieldTypeInfo{
		FieldTypeID: modula.FieldTypeID(string(d.FieldTypeID)),
		Type:        d.Type,
		Label:       d.Label,
	}
}

// fieldTypeCreateFromDb converts db CreateFieldTypeParams to SDK CreateFieldTypeParams.
func fieldTypeCreateFromDb(d db.CreateFieldTypeParams) modula.CreateFieldTypeParams {
	return modula.CreateFieldTypeParams{
		Type:  d.Type,
		Label: d.Label,
	}
}

// fieldTypeUpdateFromDb converts db UpdateFieldTypeParams to SDK UpdateFieldTypeParams.
func fieldTypeUpdateFromDb(d db.UpdateFieldTypeParams) modula.UpdateFieldTypeParams {
	return modula.UpdateFieldTypeParams{
		FieldTypeID: modula.FieldTypeID(string(d.FieldTypeID)),
		Type:        d.Type,
		Label:       d.Label,
	}
}

// ---------------------------------------------------------------------------
// Media: SDK <-> db (read-only for TUI, delete only -- no create via SDK)
// ---------------------------------------------------------------------------

// mediaToDb converts a SDK Media to a db Media.
func mediaToDb(s *modula.Media) db.Media {
	return db.Media{
		MediaID:      types.MediaID(string(s.MediaID)),
		Name:         dbNullStr(s.Name),
		DisplayName:  dbNullStr(s.DisplayName),
		Alt:          dbNullStr(s.Alt),
		Caption:      dbNullStr(s.Caption),
		Description:  dbNullStr(s.Description),
		Class:        dbNullStr(s.Class),
		Mimetype:     dbNullStr(s.Mimetype),
		Dimensions:   dbNullStr(s.Dimensions),
		URL:          types.URL(string(s.URL)),
		Srcset:       dbNullStr(s.Srcset),
		FocalX:       nullFloat64(s.FocalX),
		FocalY:       nullFloat64(s.FocalY),
		AuthorID:     nullUserID(s.AuthorID),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// mediaFromDb converts a db Media to a SDK Media.
func mediaFromDb(d db.Media) modula.Media {
	return modula.Media{
		MediaID:      modula.MediaID(string(d.MediaID)),
		Name:         dbStrPtr(d.Name),
		DisplayName:  dbStrPtr(d.DisplayName),
		Alt:          dbStrPtr(d.Alt),
		Caption:      dbStrPtr(d.Caption),
		Description:  dbStrPtr(d.Description),
		Class:        dbStrPtr(d.Class),
		Mimetype:     dbStrPtr(d.Mimetype),
		Dimensions:   dbStrPtr(d.Dimensions),
		URL:          modula.URL(string(d.URL)),
		Srcset:       dbStrPtr(d.Srcset),
		FocalX:       float64Ptr(d.FocalX),
		FocalY:       float64Ptr(d.FocalY),
		AuthorID:     userIDPtr(d.AuthorID),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// ---------------------------------------------------------------------------
// MediaDimension: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// mediaDimensionToDb converts a SDK MediaDimension to a db MediaDimensions.
func mediaDimensionToDb(s *modula.MediaDimension) db.MediaDimensions {
	return db.MediaDimensions{
		MdID:        string(s.MdID),
		Label:       dbNullStr(s.Label),
		Width:       nullInt64(s.Width),
		Height:      nullInt64(s.Height),
		AspectRatio: dbNullStr(s.AspectRatio),
	}
}

// mediaDimensionFromDb converts a db MediaDimensions to a SDK MediaDimension.
func mediaDimensionFromDb(d db.MediaDimensions) modula.MediaDimension {
	return modula.MediaDimension{
		MdID:        modula.MediaDimensionID(d.MdID),
		Label:       dbStrPtr(d.Label),
		Width:       int64Ptr(d.Width),
		Height:      int64Ptr(d.Height),
		AspectRatio: dbStrPtr(d.AspectRatio),
	}
}
