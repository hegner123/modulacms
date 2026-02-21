package db

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// nullableAdminContentIDStringEmpty returns "" when the nullable ID is invalid,
// and the ID string when valid. Used for sibling pointer fields that the
// tree builder checks with == "".
func nullableAdminContentIDStringEmpty(n types.NullableAdminContentID) string {
	if !n.Valid {
		return ""
	}
	return n.ID.String()
}

// MapAdminContentDataJSON converts AdminContentData to ContentDataJSON for tree building.
// Maps admin IDs into the public ContentDataJSON shape so BuildNodes works unchanged.
func MapAdminContentDataJSON(a AdminContentData) ContentDataJSON {
	return ContentDataJSON{
		ContentDataID: a.AdminContentDataID.String(),
		ParentID:      a.ParentID.String(),
		FirstChildID:  nullableAdminContentIDStringEmpty(a.FirstChildID),
		NextSiblingID: nullableAdminContentIDStringEmpty(a.NextSiblingID),
		PrevSiblingID: nullableAdminContentIDStringEmpty(a.PrevSiblingID),
		RouteID:       a.AdminRouteID.String(),
		DatatypeID:    a.AdminDatatypeID.String(),
		AuthorID:      a.AuthorID.String(),
		Status:        string(a.Status),
		DateCreated:   a.DateCreated.String(),
		DateModified:  a.DateModified.String(),
	}
}
