package db

import (
	"encoding/json"

	utility "github.com/hegner123/modulacms/internal/Utility"
)


//Struct Methods

func (a AdminContentData) GetHistory() string {
	return a.History.String
}

func (a AdminContentFields) GetHistory() string {
	return a.History.String
}

func (a AdminDatatypes) GetHistory() string {
	return a.History.String
}

func (a AdminFields) GetHistory() string {
	return a.History.String
}

func (a AdminRoutes) GetHistory() string {
	return a.History.String
}

func (c ContentData) GetHistory() string {
	return c.History.String
}

func (c ContentFields) GetHistory() string {
	return c.History.String
}

func (d Datatypes) GetHistory() string {
	return d.History.String
}

func (f Fields) GetHistory() string {
	return f.History.String
}

func (r Routes) GetHistory() string {
	return r.History.String
}

func (a AdminContentData) MapHistoryEntry() string {
	entry := AdminContentDataHistoryEntry{
		AdminContentDataID: a.AdminContentDataID,
		AdminRouteID:       a.AdminRouteID,
		AdminDatatypeID:    a.AdminDatatypeID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminContentFields) MapHistoryEntry() string {
	entry := AdminContentFieldsHistoryEntry{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminRouteID:        a.AdminRouteID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminDatatypes) MapHistoryEntry() string {
	entry := AdminDatatypesHistoryEntry{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminFields) MapHistoryEntry() string {
	entry := AdminFieldsHistoryEntry{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Data:         a.Data,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (a AdminRoutes) MapHistoryEntry() string {
	entry := AdminRoutesHistoryEntry{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (c ContentData) MapHistoryEntry() string {
	entry := ContentDataHistoryEntry{
		ContentDataID: c.ContentDataID,
		RouteID:       c.RouteID,
		DatatypeID:    c.DatatypeID,
		DateCreated:   c.DateCreated,
		DateModified:  c.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (c ContentFields) MapHistoryEntry() string {
	entry := ContentFieldsHistoryEntry{
		ContentFieldID: c.ContentFieldID,
		RouteID:        c.RouteID,
		ContentDataID:  c.ContentDataID,
		FieldID:        c.FieldID,
		FieldValue:     c.FieldValue,
		DateCreated:    c.DateCreated,
		DateModified:   c.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (d Datatypes) MapHistoryEntry() string {
	entry := DatatypesHistoryEntry{
		DatatypeID:   d.DatatypeID,
		ParentID:     d.ParentID,
		Label:        d.Label,
		Type:         d.Type,
		AuthorID:     d.AuthorID,
		DateCreated:  d.DateCreated,
		DateModified: d.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (f Fields) MapHistoryEntry() string {
	entry := FieldsHistoryEntry{
		FieldID:      f.FieldID,
		ParentID:     f.ParentID,
		Label:        f.Label,
		Data:         f.Data,
		Type:         f.Type,
		AuthorID:     f.AuthorID,
		DateCreated:  f.DateCreated,
		DateModified: f.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}

func (r Routes) MapHistoryEntry() string {
	entry := RoutesHistoryEntry{
		RouteID:      r.RouteID,
		Slug:         r.Slug,
		Title:        r.Title,
		Status:       r.Status,
		AuthorID:     r.AuthorID,
		DateCreated:  r.DateCreated,
		DateModified: r.DateModified,
	}
	j, err := json.Marshal(entry)
	if err != nil {
		utility.LogError("", err)
	}
	return string(j)
}
