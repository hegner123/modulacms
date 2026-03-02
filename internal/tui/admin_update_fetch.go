package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UpdateAdminFetch handles fetch messages for admin CMS entities.
func (m Model) UpdateAdminFetch(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	// =========================================================================
	// ADMIN ROUTES
	// =========================================================================
	case AdminRoutesFetchMsg:
		d := m.DB
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
		return m, func() tea.Msg {
			routes, err := d.ListAdminRoutes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return AdminRoutesFetchResultsMsg{Data: []db.AdminRoutes{}}
			}
			return AdminRoutesFetchResultsMsg{Data: *routes}
		}

	case AdminRoutesFetchResultsMsg:
		return m, tea.Batch(
			AdminRoutesSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	// =========================================================================
	// ADMIN DATATYPES
	// =========================================================================
	case AdminAllDatatypesFetchMsg:
		d := m.DB
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
		return m, func() tea.Msg {
			datatypes, err := d.ListAdminDatatypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if datatypes == nil {
				return AdminAllDatatypesFetchResultsMsg{Data: []db.AdminDatatypes{}}
			}
			return AdminAllDatatypesFetchResultsMsg{Data: *datatypes}
		}

	case AdminAllDatatypesFetchResultsMsg:
		cmds := []tea.Cmd{
			AdminAllDatatypesSetCmd(msg.Data),
			LoadingStopCmd(),
		}
		// Fetch fields for the first datatype (cursor position 0)
		if len(msg.Data) > 0 {
			cmds = append(cmds, AdminDatatypeFieldsFetchCmd(msg.Data[0].AdminDatatypeID))
		}
		return m, tea.Batch(cmds...)

	case AdminDatatypeFieldsFetchMsg:
		d := m.DB
		datatypeID := msg.AdminDatatypeID
		return m, func() tea.Msg {
			// List admin fields by parent datatype ID
			fields, err := d.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: datatypeID, Valid: true})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fields == nil || len(*fields) == 0 {
				return AdminDatatypeFieldsFetchResultsMsg{Fields: []db.AdminFields{}}
			}
			return AdminDatatypeFieldsFetchResultsMsg{Fields: *fields}
		}

	case AdminDatatypeFieldsFetchResultsMsg:
		return m, AdminDatatypeFieldsSetCmd(msg.Fields)

	// =========================================================================
	// ADMIN CONTENT DATA
	// =========================================================================
	case AdminContentDataFetchMsg:
		d := m.DB
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
		return m, func() tea.Msg {
			contentData, err := d.ListAdminContentDataTopLevelPaginated(db.PaginationParams{
				Limit:  10000,
				Offset: 0,
			})
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if contentData == nil {
				return AdminContentDataFetchResultsMsg{Data: []db.AdminContentDataTopLevel{}}
			}
			return AdminContentDataFetchResultsMsg{Data: *contentData}
		}

	case AdminContentDataFetchResultsMsg:
		return m, tea.Batch(
			AdminContentDataSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	// =========================================================================
	// FIELD TYPES
	// =========================================================================
	case FieldTypesFetchMsg:
		d := m.DB
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
		return m, func() tea.Msg {
			fieldTypes, err := d.ListFieldTypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if fieldTypes == nil {
				return FieldTypesFetchResultsMsg{Data: []db.FieldTypes{}}
			}
			return FieldTypesFetchResultsMsg{Data: *fieldTypes}
		}

	case FieldTypesFetchResultsMsg:
		return m, tea.Batch(
			FieldTypesSetCmd(msg.Data),
			LoadingStopCmd(),
		)

	// =========================================================================
	// ADMIN FIELD TYPES
	// =========================================================================
	case AdminFieldTypesFetchMsg:
		d := m.DB
		if d == nil {
			return m, func() tea.Msg {
				return FetchErrMsg{Error: fmt.Errorf("database not connected")}
			}
		}
		return m, func() tea.Msg {
			adminFieldTypes, err := d.ListAdminFieldTypes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if adminFieldTypes == nil {
				return AdminFieldTypesFetchResultsMsg{Data: []db.AdminFieldTypes{}}
			}
			return AdminFieldTypesFetchResultsMsg{Data: *adminFieldTypes}
		}

	case AdminFieldTypesFetchResultsMsg:
		return m, tea.Batch(
			AdminFieldTypesSetCmd(msg.Data),
			LoadingStopCmd(),
		)
	}

	return m, nil
}
