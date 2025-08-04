package cli

import "github.com/hegner123/modulacms/internal/db"

type ListContentDataMsg struct {}

type BuildTreeFromRouteMsg struct {
	RouteID int64
}

type ProcessContentDataRowsFromRoute struct {
	Rows []db.ContentData
}
type ProcessContentFieldRowsFromRoute struct {
	Rows []db.ContentFields
}
type ProcessDatatypeFromContentDataID struct {
	Row db.Datatypes
}
type ProcessFieldsFromContentFieldID struct {
	Rows []db.Fields
}

type GetDatatypeFromContentData struct {
	DatatypeID int64
}

type ListFieldsFromContentFields struct {
	FieldID int64
}
