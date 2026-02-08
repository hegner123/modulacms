package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tui"
)

// AdminRoutesControls handles keyboard navigation for the admin routes page.
func (m Model) AdminRoutesControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionNextPanel) {
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.AdminRoutes)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionNew) {
			return m, ShowRouteFormDialogCmd(FORMDIALOGCREATEADMINROUTE, "New Admin Route")
		}
		if km.Matches(key, config.ActionEdit) {
			if len(m.AdminRoutes) > 0 && m.Cursor < len(m.AdminRoutes) {
				route := m.AdminRoutes[m.Cursor]
				return m, ShowEditAdminRouteDialogCmd(route)
			}
		}
		if km.Matches(key, config.ActionDelete) {
			if len(m.AdminRoutes) > 0 && m.Cursor < len(m.AdminRoutes) {
				route := m.AdminRoutes[m.Cursor]
				return m, ShowDeleteAdminRouteDialogCmd(route.AdminRouteID, route.Title)
			}
		}
		if km.Matches(key, config.ActionTitlePrev) {
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		}
		if km.Matches(key, config.ActionTitleNext) {
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}
	}
	return m, nil
}

// ShowEditAdminRouteDialogCmd shows an edit dialog for an admin route.
func ShowEditAdminRouteDialogCmd(route db.AdminRoutes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditAdminRouteDialogMsg{
			Route: route,
		}
	}
}

// AdminDatatypesControls handles keyboard navigation for the admin datatypes page.
// Panel-aware: TreePanel lists datatypes, ContentPanel lists fields.
func (m Model) AdminDatatypesControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionDismiss) {
			if m.PanelFocus != tui.TreePanel {
				m.PanelFocus = tui.TreePanel
				return m, nil
			}
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionQuit) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionNextPanel) {
			m.PanelFocus = (m.PanelFocus + 1) % 3
			if m.PanelFocus == tui.ContentPanel {
				m.AdminFieldCursor = 0
			}
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			if m.PanelFocus == tui.ContentPanel {
				m.AdminFieldCursor = 0
			}
			return m, nil
		}
		if km.Matches(key, config.ActionUp) {
			return m.adminDatatypesControlsUp()
		}
		if km.Matches(key, config.ActionDown) {
			return m.adminDatatypesControlsDown()
		}
		if km.Matches(key, config.ActionBack) {
			if m.PanelFocus != tui.TreePanel {
				m.PanelFocus = tui.TreePanel
				return m, nil
			}
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if key == "l" || key == "right" {
			if m.PanelFocus == tui.TreePanel {
				m.PanelFocus = tui.ContentPanel
				m.AdminFieldCursor = 0
				return m, nil
			}
			if m.PanelFocus == tui.ContentPanel {
				m.PanelFocus = tui.RoutePanel
				return m, nil
			}
		}
		if km.Matches(key, config.ActionNew) {
			return m.adminDatatypesControlsNew()
		}
		if km.Matches(key, config.ActionEdit) {
			return m.adminDatatypesControlsEdit()
		}
		if km.Matches(key, config.ActionDelete) {
			return m.adminDatatypesControlsDelete()
		}
		if key == "enter" {
			return m.adminDatatypesControlsSelect()
		}
		if km.Matches(key, config.ActionTitlePrev) {
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		}
		if km.Matches(key, config.ActionTitleNext) {
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}
	}
	return m, nil
}

func (m Model) adminDatatypesControlsUp() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		if m.Cursor > 0 {
			newCursor := m.Cursor - 1
			if newCursor < len(m.AdminAllDatatypes) {
				dt := m.AdminAllDatatypes[newCursor]
				m.AdminFieldCursor = 0
				return m, tea.Batch(CursorUpCmd(), AdminDatatypeFieldsFetchCmd(dt.AdminDatatypeID))
			}
			return m, CursorUpCmd()
		}
	case tui.ContentPanel:
		if m.AdminFieldCursor > 0 {
			m.AdminFieldCursor--
		}
		return m, nil
	}
	return m, nil
}

func (m Model) adminDatatypesControlsDown() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		if m.Cursor < len(m.AdminAllDatatypes)-1 {
			newCursor := m.Cursor + 1
			if newCursor < len(m.AdminAllDatatypes) {
				dt := m.AdminAllDatatypes[newCursor]
				m.AdminFieldCursor = 0
				return m, tea.Batch(CursorDownCmd(), AdminDatatypeFieldsFetchCmd(dt.AdminDatatypeID))
			}
			return m, CursorDownCmd()
		}
	case tui.ContentPanel:
		if m.AdminFieldCursor < len(m.AdminSelectedDatatypeFields)-1 {
			m.AdminFieldCursor++
		}
		return m, nil
	}
	return m, nil
}

func (m Model) adminDatatypesControlsNew() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		return m, ShowAdminFormDialogCmd(FORMDIALOGCREATEADMINDATATYPE, "New Admin Datatype", m.AdminAllDatatypes)
	case tui.ContentPanel:
		if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
			return m, ShowFieldFormDialogCmd(FORMDIALOGCREATEADMINFIELD, "New Admin Field")
		}
	}
	return m, nil
}

func (m Model) adminDatatypesControlsEdit() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
			dt := m.AdminAllDatatypes[m.Cursor]
			return m, ShowEditAdminDatatypeDialogCmd(dt, m.AdminAllDatatypes)
		}
	case tui.ContentPanel:
		if len(m.AdminSelectedDatatypeFields) > 0 && m.AdminFieldCursor < len(m.AdminSelectedDatatypeFields) {
			field := m.AdminSelectedDatatypeFields[m.AdminFieldCursor]
			return m, ShowEditAdminFieldDialogCmd(field)
		}
	}
	return m, nil
}

func (m Model) adminDatatypesControlsDelete() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
			dt := m.AdminAllDatatypes[m.Cursor]
			hasChildren := false
			for _, other := range m.AdminAllDatatypes {
				if other.ParentID.Valid && string(other.ParentID.ID) == string(dt.AdminDatatypeID) {
					hasChildren = true
					break
				}
			}
			return m, ShowDeleteAdminDatatypeDialogCmd(dt.AdminDatatypeID, dt.Label, hasChildren)
		}
	case tui.ContentPanel:
		if len(m.AdminSelectedDatatypeFields) > 0 && m.AdminFieldCursor < len(m.AdminSelectedDatatypeFields) {
			field := m.AdminSelectedDatatypeFields[m.AdminFieldCursor]
			var datatypeID types.AdminDatatypeID
			if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
				datatypeID = m.AdminAllDatatypes[m.Cursor].AdminDatatypeID
			}
			return m, ShowDeleteAdminFieldDialogCmd(field.AdminFieldID, datatypeID, field.Label)
		}
	}
	return m, nil
}

func (m Model) adminDatatypesControlsSelect() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		if len(m.AdminAllDatatypes) > 0 && m.Cursor < len(m.AdminAllDatatypes) {
			dt := m.AdminAllDatatypes[m.Cursor]
			m.PanelFocus = tui.ContentPanel
			m.AdminFieldCursor = 0
			return m, LogMessageCmd(fmt.Sprintf("Admin datatype selected: %s (%s)", dt.Label, dt.AdminDatatypeID))
		}
	case tui.ContentPanel:
		if len(m.AdminSelectedDatatypeFields) > 0 && m.AdminFieldCursor < len(m.AdminSelectedDatatypeFields) {
			field := m.AdminSelectedDatatypeFields[m.AdminFieldCursor]
			return m, LogMessageCmd(fmt.Sprintf("Admin field selected: %s [%s]", field.Label, field.Type))
		}
	}
	return m, nil
}

// ShowAdminFormDialogCmd shows a form dialog for admin datatype create with parent options from admin datatypes.
func ShowAdminFormDialogCmd(action FormDialogAction, title string, parents []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowAdminFormDialogMsg{
			Action:  action,
			Title:   title,
			Parents: parents,
		}
	}
}

// ShowEditAdminDatatypeDialogCmd shows an edit dialog for an admin datatype.
func ShowEditAdminDatatypeDialogCmd(dt db.AdminDatatypes, allDatatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditAdminDatatypeDialogMsg{
			Datatype: dt,
			Parents:  allDatatypes,
		}
	}
}

// ShowEditAdminFieldDialogCmd shows an edit dialog for an admin field.
func ShowEditAdminFieldDialogCmd(field db.AdminFields) tea.Cmd {
	return func() tea.Msg {
		return ShowEditAdminFieldDialogMsg{
			Field: field,
		}
	}
}

// AdminContentBrowserControls handles keyboard navigation for admin content page.
// Simplified flat list of m.AdminRootContentSummary.
func (m Model) AdminContentBrowserControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionNextPanel) {
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.AdminRootContentSummary)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionDelete) {
			if len(m.AdminRootContentSummary) > 0 && m.Cursor < len(m.AdminRootContentSummary) {
				content := m.AdminRootContentSummary[m.Cursor]
				return m, DeleteAdminContentCmd(content.AdminContentDataID)
			}
		}
		if km.Matches(key, config.ActionTitlePrev) {
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		}
		if km.Matches(key, config.ActionTitleNext) {
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}
	}
	return m, nil
}
