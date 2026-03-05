package remote

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// ContentData: SDK <-> db
// ---------------------------------------------------------------------------

// contentDataToDb converts a SDK ContentData to a db ContentData.
func contentDataToDb(s *modula.ContentData) db.ContentData {
	return db.ContentData{
		ContentDataID: types.ContentID(string(s.ContentDataID)),
		ParentID:      nullContentID(s.ParentID),
		FirstChildID:  nullContentIDFromString(s.FirstChildID),
		NextSiblingID: nullContentIDFromString(s.NextSiblingID),
		PrevSiblingID: nullContentIDFromString(s.PrevSiblingID),
		RouteID:       nullRouteID(s.RouteID),
		DatatypeID:    nullDatatypeID(s.DatatypeID),
		AuthorID:      userIDPtrToDb(s.AuthorID),
		Status:        types.ContentStatus(string(s.Status)),
		PublishedAt:   sdkTimestampPtrToDb(s.PublishedAt),
		PublishedBy:   nullUserID(s.PublishedBy),
		PublishAt:     sdkTimestampPtrToDb(s.PublishAt),
		Revision:      s.Revision,
		DateCreated:   sdkTimestampToDb(s.DateCreated),
		DateModified:  sdkTimestampToDb(s.DateModified),
	}
}

// contentDataFromDb converts a db ContentData to a SDK ContentData.
func contentDataFromDb(d db.ContentData) modula.ContentData {
	var parentID *modula.ContentID
	if d.ParentID.Valid {
		id := modula.ContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.ContentData{
		ContentDataID: modula.ContentID(string(d.ContentDataID)),
		ParentID:      parentID,
		FirstChildID:  nullableContentIDToString(d.FirstChildID),
		NextSiblingID: nullableContentIDToString(d.NextSiblingID),
		PrevSiblingID: nullableContentIDToString(d.PrevSiblingID),
		RouteID:       routeIDPtr(d.RouteID),
		DatatypeID:    datatypeIDPtr(d.DatatypeID),
		AuthorID:      userIDToSdkPtr(d.AuthorID),
		Status:        modula.ContentStatus(string(d.Status)),
		PublishedAt:   dbTimestampToSdkPtr(d.PublishedAt),
		PublishedBy:   userIDPtr(d.PublishedBy),
		PublishAt:     dbTimestampToSdkPtr(d.PublishAt),
		Revision:      d.Revision,
		DateCreated:   dbTimestampToSdk(d.DateCreated),
		DateModified:  dbTimestampToSdk(d.DateModified),
	}
}

// contentDataCreateFromDb converts db CreateContentDataParams to SDK CreateContentDataParams.
func contentDataCreateFromDb(d db.CreateContentDataParams) modula.CreateContentDataParams {
	var parentID *modula.ContentID
	if d.ParentID.Valid {
		id := modula.ContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.CreateContentDataParams{
		ParentID:      parentID,
		FirstChildID:  nullableContentIDToString(d.FirstChildID),
		NextSiblingID: nullableContentIDToString(d.NextSiblingID),
		PrevSiblingID: nullableContentIDToString(d.PrevSiblingID),
		RouteID:       routeIDPtr(d.RouteID),
		DatatypeID:    datatypeIDPtr(d.DatatypeID),
		AuthorID:      userIDToSdkPtr(d.AuthorID),
		Status:        modula.ContentStatus(string(d.Status)),
	}
}

// contentDataUpdateFromDb converts db UpdateContentDataParams to SDK UpdateContentDataParams.
func contentDataUpdateFromDb(d db.UpdateContentDataParams) modula.UpdateContentDataParams {
	var parentID *modula.ContentID
	if d.ParentID.Valid {
		id := modula.ContentID(string(d.ParentID.ID))
		parentID = &id
	}
	return modula.UpdateContentDataParams{
		ContentDataID: modula.ContentID(string(d.ContentDataID)),
		ParentID:      parentID,
		FirstChildID:  nullableContentIDToString(d.FirstChildID),
		NextSiblingID: nullableContentIDToString(d.NextSiblingID),
		PrevSiblingID: nullableContentIDToString(d.PrevSiblingID),
		RouteID:       routeIDPtr(d.RouteID),
		DatatypeID:    datatypeIDPtr(d.DatatypeID),
		AuthorID:      userIDToSdkPtr(d.AuthorID),
		Status:        modula.ContentStatus(string(d.Status)),
	}
}

// ---------------------------------------------------------------------------
// ContentField: SDK <-> db
// ---------------------------------------------------------------------------

// contentFieldToDb converts a SDK ContentField to a db ContentFields.
func contentFieldToDb(s *modula.ContentField) db.ContentFields {
	return db.ContentFields{
		ContentFieldID: types.ContentFieldID(string(s.ContentFieldID)),
		RouteID:        nullRouteID(s.RouteID),
		ContentDataID:  nullContentID(s.ContentDataID),
		FieldID:        nullFieldID(s.FieldID),
		FieldValue:     s.FieldValue,
		Locale:         s.Locale,
		AuthorID:       userIDPtrToDb(s.AuthorID),
		DateCreated:    sdkTimestampToDb(s.DateCreated),
		DateModified:   sdkTimestampToDb(s.DateModified),
	}
}

// contentFieldFromDb converts a db ContentFields to a SDK ContentField.
func contentFieldFromDb(d db.ContentFields) modula.ContentField {
	return modula.ContentField{
		ContentFieldID: modula.ContentFieldID(string(d.ContentFieldID)),
		RouteID:        routeIDPtr(d.RouteID),
		ContentDataID:  contentIDPtr(d.ContentDataID),
		FieldID:        fieldIDPtr(d.FieldID),
		FieldValue:     d.FieldValue,
		Locale:         d.Locale,
		AuthorID:       userIDToSdkPtr(d.AuthorID),
		DateCreated:    dbTimestampToSdk(d.DateCreated),
		DateModified:   dbTimestampToSdk(d.DateModified),
	}
}

// contentFieldCreateFromDb converts db CreateContentFieldParams to SDK CreateContentFieldParams.
func contentFieldCreateFromDb(d db.CreateContentFieldParams) modula.CreateContentFieldParams {
	return modula.CreateContentFieldParams{
		RouteID:       routeIDPtr(d.RouteID),
		ContentDataID: contentIDPtr(d.ContentDataID),
		FieldID:       fieldIDPtr(d.FieldID),
		FieldValue:    d.FieldValue,
		AuthorID:      userIDToSdkPtr(d.AuthorID),
	}
}

// contentFieldUpdateFromDb converts db UpdateContentFieldParams to SDK UpdateContentFieldParams.
func contentFieldUpdateFromDb(d db.UpdateContentFieldParams) modula.UpdateContentFieldParams {
	return modula.UpdateContentFieldParams{
		ContentFieldID: modula.ContentFieldID(string(d.ContentFieldID)),
		RouteID:        routeIDPtr(d.RouteID),
		ContentDataID:  contentIDPtr(d.ContentDataID),
		FieldID:        fieldIDPtr(d.FieldID),
		FieldValue:     d.FieldValue,
		AuthorID:       userIDToSdkPtr(d.AuthorID),
	}
}

// ---------------------------------------------------------------------------
// ContentRelation: SDK <-> db (read-only for TUI; no create/update needed via remote)
// ---------------------------------------------------------------------------

// contentRelationToDb converts a SDK ContentRelation to a db ContentRelations.
func contentRelationToDb(s *modula.ContentRelation) db.ContentRelations {
	return db.ContentRelations{
		ContentRelationID: types.ContentRelationID(string(s.ContentRelationID)),
		SourceContentID:   types.ContentID(string(s.SourceContentID)),
		TargetContentID:   types.ContentID(string(s.TargetContentID)),
		FieldID:           types.FieldID(string(s.FieldID)),
		SortOrder:         s.SortOrder,
		DateCreated:       sdkTimestampToDb(s.DateCreated),
	}
}

// contentRelationFromDb converts a db ContentRelations to a SDK ContentRelation.
func contentRelationFromDb(d db.ContentRelations) modula.ContentRelation {
	return modula.ContentRelation{
		ContentRelationID: modula.ContentRelationID(string(d.ContentRelationID)),
		SourceContentID:   modula.ContentID(string(d.SourceContentID)),
		TargetContentID:   modula.ContentID(string(d.TargetContentID)),
		FieldID:           modula.FieldID(string(d.FieldID)),
		SortOrder:         d.SortOrder,
		DateCreated:       dbTimestampToSdk(d.DateCreated),
	}
}

// ---------------------------------------------------------------------------
// ContentVersion: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// contentVersionToDb converts a SDK ContentVersion to a db ContentVersion.
func contentVersionToDb(s *modula.ContentVersion) db.ContentVersion {
	return db.ContentVersion{
		ContentVersionID: types.ContentVersionID(string(s.ContentVersionID)),
		ContentDataID:    types.ContentID(string(s.ContentDataID)),
		VersionNumber:    s.VersionNumber,
		Locale:           s.Locale,
		Snapshot:         s.Snapshot,
		Trigger:          s.Trigger,
		Label:            s.Label,
		Published:        s.Published,
		PublishedBy:      nullUserID(s.PublishedBy),
		DateCreated:      sdkTimestampToDb(s.DateCreated),
	}
}

// contentVersionFromDb converts a db ContentVersion to a SDK ContentVersion.
func contentVersionFromDb(d db.ContentVersion) modula.ContentVersion {
	return modula.ContentVersion{
		ContentVersionID: modula.ContentVersionID(string(d.ContentVersionID)),
		ContentDataID:    modula.ContentID(string(d.ContentDataID)),
		VersionNumber:    d.VersionNumber,
		Locale:           d.Locale,
		Snapshot:         d.Snapshot,
		Trigger:          d.Trigger,
		Label:            d.Label,
		Published:        d.Published,
		PublishedBy:      userIDPtr(d.PublishedBy),
		DateCreated:      dbTimestampToSdk(d.DateCreated),
	}
}

// ---------------------------------------------------------------------------
// AdminContentVersion: SDK <-> db
// ---------------------------------------------------------------------------

// adminContentVersionToDb converts a SDK AdminContentVersion to a db AdminContentVersion.
func adminContentVersionToDb(s *modula.AdminContentVersion) db.AdminContentVersion {
	return db.AdminContentVersion{
		AdminContentVersionID: types.AdminContentVersionID(string(s.AdminContentVersionID)),
		AdminContentDataID:    types.AdminContentID(string(s.AdminContentDataID)),
		VersionNumber:         s.VersionNumber,
		Locale:                s.Locale,
		Snapshot:              s.Snapshot,
		Trigger:               s.Trigger,
		Label:                 s.Label,
		Published:             s.Published,
		PublishedBy:           nullUserID(s.PublishedBy),
		DateCreated:           sdkTimestampToDb(s.DateCreated),
	}
}
