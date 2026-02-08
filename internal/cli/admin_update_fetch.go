package cli

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
			// Get field IDs from the admin join table
			dtID := types.NullableAdminDatatypeID{ID: datatypeID, Valid: true}
			dtFields, err := d.ListAdminDatatypeFieldByDatatypeID(dtID)
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if dtFields == nil || len(*dtFields) == 0 {
				return AdminDatatypeFieldsFetchResultsMsg{Fields: []db.AdminFields{}}
			}

			// Fetch actual field details for each field ID
			var fields []db.AdminFields
			for _, dtf := range *dtFields {
				if dtf.AdminFieldID.Valid {
					field, err := d.GetAdminField(dtf.AdminFieldID.ID)
					if err == nil && field != nil {
						fields = append(fields, *field)
					}
				}
			}
			return AdminDatatypeFieldsFetchResultsMsg{Fields: fields}
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
			contentData, err := d.ListAdminContentData()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if contentData == nil {
				return AdminContentDataFetchResultsMsg{Data: []db.AdminContentData{}}
			}
			return AdminContentDataFetchResultsMsg{Data: *contentData}
		}

	case AdminContentDataFetchResultsMsg:
		return m, tea.Batch(
			AdminContentDataSetCmd(msg.Data),
			LoadingStopCmd(),
		)
	}

	return m, nil
}
