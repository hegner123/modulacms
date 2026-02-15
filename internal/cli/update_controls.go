package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
	"github.com/hegner123/modulacms/internal/tui"
)

// ControlUpdate signals that page-specific key handling should proceed.
type ControlUpdate struct{}

// NewControlUpdate returns a command that creates a ControlUpdate message.
func NewControlUpdate() tea.Cmd {
	return func() tea.Msg {
		return ControlUpdate{}
	}
}

// PageSpecificMsgHandlers routes keyboard events to the appropriate page handler.
func (m Model) PageSpecificMsgHandlers(cmd tea.Cmd, msg tea.Msg) (Model, tea.Cmd) {

	switch m.Page.Index {
	case HOMEPAGE:
		return m.BasicControls(msg)
	case CMSPAGE:
		return m.BasicCMSControls(msg)
	case ADMINCMSPAGE:
		return m.BasicCMSControls(msg)
	case DATABASEPAGE:
		return m.SelectTable(msg)
	case TABLEPAGE:
		return m.TablePageControls(msg)
	case DYNAMICPAGE:
		return m.BasicDynamicControls(msg)
	case READPAGE:
		return m.TableNavigationControls(msg)
	case READSINGLEPAGE:
		return m.TableNavigationControls(msg)
	case UPDATEPAGE:
		return m.TableNavigationControls(msg)
	case DELETEPAGE:
		return m.TableNavigationControls(msg)
	case DATATYPES:
		return m.DatatypesControls(msg)
	case DEVELOPMENT:
		return DevelopmentInterface(m, msg)
	case DATATYPE:
		return m.DefineDatatypeControls(msg)
	case CONFIGPAGE:
		return m.ConfigControls(msg)
	case ACTIONSPAGE:
		return m.ActionsControls(msg)
	case ROUTES:
		return m.RoutesControls(msg)
	case CONTENT:
		return m.ContentBrowserControls(msg)
	case USERSADMIN:
		return m.UsersAdminControls(msg)
	case MEDIA:
		return m.MediaControls(msg)
	case ADMINROUTES:
		return m.AdminRoutesControls(msg)
	case ADMINDATATYPES:
		return m.AdminDatatypesControls(msg)
	case ADMINCONTENT:
		return m.AdminContentBrowserControls(msg)

	}
	return m, nil
}

// BasicControls handles keyboard navigation for simple menu-based pages.
func (m Model) BasicControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
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
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.PageMenu)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			if len(m.PageMenu) > 0 {
				return m, NavigateToPageCmd(m.PageMenu[m.Cursor])
			}
		}
	}
	return m, nil
}

// TablePageControls handles TABLEPAGE key events, intercepting CREATE to show
// the database insert dialog instead of navigating to the old form page.
func (m Model) TablePageControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
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
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.PageMenu)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			if len(m.PageMenu) > 0 {
				selected := m.PageMenu[m.Cursor]
				if selected.Index == CREATEPAGE {
					// Intercept: show database insert dialog instead of navigating
					return m, ShowDatabaseInsertDialogCmd(db.DBTable(m.TableState.Table))
				}
				return m, NavigateToPageCmd(selected)
			}
		}
	}
	return m, nil
}

// BasicCMSControls handles keyboard navigation for the CMS main page with 3-panel layout.
func (m Model) BasicCMSControls(msg tea.Msg) (Model, tea.Cmd) {
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
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.PageMenu) {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			page := m.PageMenu[m.Cursor]
			switch page.Index {
			case ROUTES:
				return m, NavigateToPageCmd(m.PageMap[ROUTES])
			case DATATYPES:
				return m, NavigateToPageCmd(m.PageMap[DATATYPES])
			case FIELDS:
				return m, tea.Batch()
			case CONTENT:
				return m, NavigateToPageCmd(m.PageMap[CONTENT])
			case MEDIA:
				return m, NavigateToPageCmd(m.PageMap[MEDIA])
			case USERSADMIN:
				return m, NavigateToPageCmd(m.PageMap[USERSADMIN])
			case ADMINROUTES:
				return m, NavigateToPageCmd(m.PageMap[ADMINROUTES])
			case ADMINDATATYPES:
				return m, NavigateToPageCmd(m.PageMap[ADMINDATATYPES])
			case ADMINCONTENT:
				return m, NavigateToPageCmd(m.PageMap[ADMINCONTENT])
			default:
				return m, nil
			}
		}
	}
	return m, nil
}

// RoutesControls handles keyboard navigation for the routes page
func (m Model) RoutesControls(msg tea.Msg) (Model, tea.Cmd) {
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
			if m.Cursor < len(m.Routes)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
				route := m.Routes[m.Cursor]
				m.PageRouteId = route.RouteID
				return m, LogMessageCmd(fmt.Sprintf("Route selected: %s (%s)", route.Title, route.RouteID))
			}
		}
		if km.Matches(key, config.ActionNew) {
			return m, ShowRouteFormDialogCmd(FORMDIALOGCREATEROUTE, "New Route")
		}
		if km.Matches(key, config.ActionEdit) {
			if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
				return m, ShowEditRouteDialogCmd(m.Routes[m.Cursor])
			}
		}
		if km.Matches(key, config.ActionDelete) {
			if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
				route := m.Routes[m.Cursor]
				return m, ShowDeleteRouteDialogCmd(route.RouteID, route.Title)
			}
		}
	}
	return m, nil
}

// DatatypesControls handles keyboard navigation for the datatypes panel page.
// Panel-specific controls:
//   - TreePanel (left): Navigate datatypes list, 'n' creates new datatype
//   - ContentPanel (center): Navigate fields list, 'n' creates new field
//   - RoutePanel (right): Actions panel (info only)
// DatatypesControls handles keyboard navigation for the datatypes page.
func (m Model) DatatypesControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		// Dismiss first: esc goes back to tree panel before quitting
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
				m.FieldCursor = 0
			}
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			if m.PanelFocus == tui.ContentPanel {
				m.FieldCursor = 0
			}
			return m, nil
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
		if km.Matches(key, config.ActionUp) {
			return m.datatypesControlsUp()
		}
		if km.Matches(key, config.ActionDown) {
			return m.datatypesControlsDown()
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
		// Right arrow moves to next panel (subset of select behavior)
		if key == "l" || key == "right" {
			if m.PanelFocus == tui.TreePanel {
				m.PanelFocus = tui.ContentPanel
				m.FieldCursor = 0
				return m, nil
			}
			if m.PanelFocus == tui.ContentPanel {
				m.PanelFocus = tui.RoutePanel
				return m, nil
			}
		}
		if km.Matches(key, config.ActionNew) {
			return m.datatypesControlsNew()
		}
		if km.Matches(key, config.ActionEdit) {
			return m.datatypesControlsEdit()
		}
		if km.Matches(key, config.ActionDelete) {
			return m.datatypesControlsDelete()
		}
		if key == "enter" {
			return m.datatypesControlsSelect()
		}
	}
	return m, nil
}

// datatypesControlsUp handles upward cursor movement based on active panel.
func (m Model) datatypesControlsUp() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Navigate datatypes list
		if m.Cursor > 0 {
			newCursor := m.Cursor - 1
			if newCursor < len(m.AllDatatypes) {
				dt := m.AllDatatypes[newCursor]
				m.FieldCursor = 0 // Reset field cursor when changing datatype
				return m, tea.Batch(CursorUpCmd(), DatatypeFieldsFetchCmd(dt.DatatypeID))
			}
			return m, CursorUpCmd()
		}
	case tui.ContentPanel:
		// Navigate fields list
		if m.FieldCursor > 0 {
			m.FieldCursor--
		}
		return m, nil
	}
	return m, nil
}

// datatypesControlsDown handles downward cursor movement based on active panel.
func (m Model) datatypesControlsDown() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Navigate datatypes list
		if m.Cursor < len(m.AllDatatypes)-1 {
			newCursor := m.Cursor + 1
			if newCursor < len(m.AllDatatypes) {
				dt := m.AllDatatypes[newCursor]
				m.FieldCursor = 0 // Reset field cursor when changing datatype
				return m, tea.Batch(CursorDownCmd(), DatatypeFieldsFetchCmd(dt.DatatypeID))
			}
			return m, CursorDownCmd()
		}
	case tui.ContentPanel:
		// Navigate fields list
		maxFields := len(m.SelectedDatatypeFields)
		if m.FieldCursor < maxFields-1 {
			m.FieldCursor++
		}
		return m, nil
	}
	return m, nil
}

// datatypesControlsNew handles creation key based on active panel.
func (m Model) datatypesControlsNew() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Create new datatype
		return m, CmsDefineDatatypeLoadCmd()
	case tui.ContentPanel:
		// Create new field for the selected datatype
		if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
			return m, ShowFieldFormDialogCmd(FORMDIALOGCREATEFIELD, "New Field")
		}
	}
	return m, nil
}

// datatypesControlsEdit handles edit key based on active panel.
func (m Model) datatypesControlsEdit() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Edit selected datatype using modal dialog
		if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
			return m, ShowEditDatatypeDialogCmd(m.AllDatatypes[m.Cursor], m.AllDatatypes)
		}
	case tui.ContentPanel:
		// Edit selected field using modal dialog
		if len(m.SelectedDatatypeFields) > 0 && m.FieldCursor < len(m.SelectedDatatypeFields) {
			field := m.SelectedDatatypeFields[m.FieldCursor]
			return m, ShowEditFieldDialogCmd(field)
		}
	}
	return m, nil
}

// datatypesControlsDelete handles deletion key based on active panel.
func (m Model) datatypesControlsDelete() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Delete selected datatype
		if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
			dt := m.AllDatatypes[m.Cursor]
			// Check if any other datatype has this as parent
			hasChildren := false
			for _, other := range m.AllDatatypes {
				if other.ParentID.Valid && types.DatatypeID(other.ParentID.ID) == dt.DatatypeID {
					hasChildren = true
					break
				}
			}
			return m, ShowDeleteDatatypeDialogCmd(dt.DatatypeID, dt.Label, hasChildren)
		}
	case tui.ContentPanel:
		// Delete selected field from the datatype
		if len(m.SelectedDatatypeFields) > 0 && m.FieldCursor < len(m.SelectedDatatypeFields) {
			field := m.SelectedDatatypeFields[m.FieldCursor]
			var datatypeID types.DatatypeID
			if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
				datatypeID = m.AllDatatypes[m.Cursor].DatatypeID
			}
			return m, ShowDeleteFieldDialogCmd(field.FieldID, datatypeID, field.Label)
		}
	}
	return m, nil
}

// datatypesControlsSelect handles selection key based on active panel.
func (m Model) datatypesControlsSelect() (Model, tea.Cmd) {
	switch m.PanelFocus {
	case tui.TreePanel:
		// Select datatype (move focus to fields)
		if len(m.AllDatatypes) > 0 && m.Cursor < len(m.AllDatatypes) {
			dt := m.AllDatatypes[m.Cursor]
			m.PanelFocus = tui.ContentPanel
			m.FieldCursor = 0
			return m, LogMessageCmd(fmt.Sprintf("Datatype selected: %s (%s)", dt.Label, dt.DatatypeID))
		}
	case tui.ContentPanel:
		// Select field (show field details or edit)
		if len(m.SelectedDatatypeFields) > 0 && m.FieldCursor < len(m.SelectedDatatypeFields) {
			field := m.SelectedDatatypeFields[m.FieldCursor]
			return m, LogMessageCmd(fmt.Sprintf("Field selected: %s [%s]", field.Label, field.Type))
		}
	}
	return m, nil
}

// ContentBrowserControls handles keyboard navigation for the content browser.
// The Content page shows:
//   - Left panel: content data instances with slug and ROOT datatype label
//   - Center panel: details of selected content (or content tree if viewing)
//   - Right panel: actions
func (m Model) ContentBrowserControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		// Panel focus cycling
		if km.Matches(key, config.ActionNextPanel) {
			m.PanelFocus = (m.PanelFocus + 1) % 3
			if m.PanelFocus == tui.RoutePanel {
				m.FieldCursor = 0
			}
			return m, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			m.PanelFocus = (m.PanelFocus + 2) % 3
			if m.PanelFocus == tui.RoutePanel {
				m.FieldCursor = 0
			}
			return m, nil
		}

		// Exit: ctrl+c always quits immediately
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		// q shows quit confirmation dialog
		if km.Matches(key, config.ActionQuit) && !km.Matches(key, config.ActionDismiss) && !km.Matches(key, config.ActionBack) {
			return m, ShowQuitConfirmDialogCmd()
		}
		// Esc / back: step back through content flow states
		if km.Matches(key, config.ActionDismiss) || km.Matches(key, config.ActionBack) {
			if !m.PageRouteId.IsZero() {
				m.PageRouteId = types.RouteID("")
				m.Root = tree.Root{}
				m.Cursor = 0
				return m, nil
			}
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
			return m, tea.Quit
		}

		// Navigation
		if km.Matches(key, config.ActionUp) {
			if m.PageRouteId.IsZero() {
				if m.Cursor > 0 {
					return m, CursorUpCmd()
				}
			} else if m.PanelFocus == tui.RoutePanel {
				// Navigate fields in right panel
				if m.FieldCursor > 0 {
					m.FieldCursor--
				}
				return m, nil
			} else {
				if m.Cursor > 0 {
					return m, contentBrowserCursorUpCmd(m)
				}
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.PageRouteId.IsZero() {
				if m.Cursor < len(m.RootContentSummary)-1 {
					return m, CursorDownCmd()
				}
			} else if m.PanelFocus == tui.RoutePanel {
				// Navigate fields in right panel
				if m.FieldCursor < len(m.SelectedContentFields)-1 {
					m.FieldCursor++
				}
				return m, nil
			} else {
				maxCursor := m.Root.CountVisible()
				if m.Cursor < maxCursor-1 {
					return m, contentBrowserCursorDownCmd(m)
				}
			}
		}

		// Selection
		if km.Matches(key, config.ActionSelect) {
			if m.PageRouteId.IsZero() {
				if len(m.RootContentSummary) > 0 && m.Cursor < len(m.RootContentSummary) {
					content := m.RootContentSummary[m.Cursor]
					if content.RouteID.Valid {
						m.PageRouteId = content.RouteID.ID
						m.SelectedDatatype = content.DatatypeID.ID
						m.Cursor = 0
						return m, tea.Batch(
							LoadingStartCmd(),
							LogMessageCmd(fmt.Sprintf("Content selected: %s [%s]", content.RouteSlug, content.DatatypeLabel)),
							ReloadContentTreeCmd(m.Config, content.RouteID.ID),
						)
					}
				}
			} else {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.FirstChild != nil {
					node.Expand = !node.Expand
					return m, nil
				}
			}
		}

		// Expand/Collapse
		if km.Matches(key, config.ActionExpand) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.FirstChild != nil {
					node.Expand = true
					return m, nil
				}
			}
		}
		if km.Matches(key, config.ActionCollapse) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.FirstChild != nil {
					node.Expand = false
					return m, nil
				}
			}
		}

		// Actions
		if km.Matches(key, config.ActionEdit) {
			if !m.PageRouteId.IsZero() {
				if m.PanelFocus == tui.RoutePanel {
					return m.contentFieldEdit()
				}
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					m.Logger.Finfo(fmt.Sprintf("'e' key pressed, editing node %s with datatype %s", node.Instance.ContentDataID, node.Datatype.Label))
					return m, FetchContentForEditCmd(
						node.Instance.ContentDataID,
						node.Datatype.DatatypeID,
						m.PageRouteId,
						fmt.Sprintf("Edit: %s", node.Datatype.Label),
					)
				}
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		}
		if km.Matches(key, config.ActionNew) {
			if m.PageRouteId.IsZero() {
				if len(m.RootContentSummary) > 0 && m.Cursor < len(m.RootContentSummary) {
					content := m.RootContentSummary[m.Cursor]
					if content.DatatypeID.Valid {
						return m, ShowCreateRouteWithContentDialogCmd(string(content.DatatypeID.ID))
					}
				} else if len(m.RootDatatypes) > 0 {
					return m, ShowCreateRouteWithContentDialogCmd(string(m.RootDatatypes[0].DatatypeID))
				}
			} else {
				if m.PanelFocus == tui.RoutePanel {
					return m.contentFieldAdd()
				}
				node := m.Root.NodeAtIndex(m.Cursor)
				m.Logger.Finfo(fmt.Sprintf("'n' key pressed, node: %v", node != nil))
				if node != nil {
					m.Logger.Finfo(fmt.Sprintf("Showing child datatype picker for parent %s", node.Datatype.Label))
					return m, ShowChildDatatypeDialogCmd(node.Datatype.DatatypeID, m.PageRouteId)
				}
				m.Logger.Finfo("No node selected")
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		}
		if km.Matches(key, config.ActionDelete) {
			if !m.PageRouteId.IsZero() {
				if m.PanelFocus == tui.RoutePanel {
					return m.contentFieldDelete()
				}
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					contentName := DecideNodeName(*node)
					hasChildren := node.FirstChild != nil
					return m, ShowDeleteContentDialogCmd(
						string(node.Instance.ContentDataID),
						contentName,
						hasChildren,
					)
				}
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		}
		if km.Matches(key, config.ActionMove) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					// Build valid target list: all visible nodes except self and descendants
					allVisible := m.Root.FlattenVisible()
					targets := make([]ParentOption, 0)
					for _, candidate := range allVisible {
						if candidate.Instance == nil {
							continue
						}
						if candidate.Instance.ContentDataID == node.Instance.ContentDataID {
							continue // skip self
						}
						if tree.IsDescendantOf(candidate, node) {
							continue // skip descendants of source
						}
						label := DecideNodeName(*candidate)
						targets = append(targets, ParentOption{
							Label: label,
							Value: string(candidate.Instance.ContentDataID),
						})
					}
					if len(targets) == 0 {
						return m, ShowDialog("Cannot Move", "No valid move targets", false)
					}
					return m, ShowMoveContentDialogCmd(node, m.PageRouteId, targets)
				}
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		}
		if km.Matches(key, config.ActionReorderUp) {
			if !m.PageRouteId.IsZero() {
				if m.PanelFocus == tui.RoutePanel {
					return m.contentFieldReorderUp()
				}
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil && node.PrevSibling != nil {
					return m, tea.Batch(LoadingStartCmd(), ReorderSiblingCmd(node.Instance.ContentDataID, m.PageRouteId, "up"))
				}
			}
		}
		if km.Matches(key, config.ActionReorderDown) {
			if !m.PageRouteId.IsZero() {
				if m.PanelFocus == tui.RoutePanel {
					return m.contentFieldReorderDown()
				}
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil && node.NextSibling != nil {
					return m, tea.Batch(LoadingStartCmd(), ReorderSiblingCmd(node.Instance.ContentDataID, m.PageRouteId, "down"))
				}
			}
		}
		if km.Matches(key, config.ActionCopy) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					return m, tea.Batch(LoadingStartCmd(), CopyContentCmd(node.Instance.ContentDataID, m.PageRouteId))
				}
			}
		}
		if km.Matches(key, config.ActionPublish) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					return m, tea.Batch(LoadingStartCmd(), TogglePublishCmd(node.Instance.ContentDataID, m.PageRouteId))
				}
			}
		}
		if km.Matches(key, config.ActionArchive) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Instance != nil {
					return m, tea.Batch(LoadingStartCmd(), ArchiveContentCmd(node.Instance.ContentDataID, m.PageRouteId))
				}
			}
		}

		// Navigate to parent
		if km.Matches(key, config.ActionGoParent) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.Parent != nil && node.Parent.Instance != nil {
					idx := m.Root.FindVisibleIndex(node.Parent.Instance.ContentDataID)
					if idx >= 0 {
						m.Cursor = idx
						return m, nil
					}
				}
			}
		}
		// Navigate to first child
		if km.Matches(key, config.ActionGoChild) {
			if !m.PageRouteId.IsZero() {
				node := m.Root.NodeAtIndex(m.Cursor)
				if node != nil && node.FirstChild != nil {
					node.Expand = true
					idx := m.Root.FindVisibleIndex(node.FirstChild.Instance.ContentDataID)
					if idx >= 0 {
						m.Cursor = idx
						return m, nil
					}
				}
			}
		}

		// Title font change
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

// contentBrowserCursorUpCmd returns a command to move cursor up and load fields for the previous node.
func contentBrowserCursorUpCmd(m Model) tea.Cmd {
	newCursor := m.Cursor - 1
	node := m.Root.NodeAtIndex(newCursor)
	if node != nil && node.Instance != nil {
		return tea.Batch(CursorUpCmd(), LoadContentFieldsCmd(m.Config, node.Instance.ContentDataID, node.Instance.DatatypeID))
	}
	return CursorUpCmd()
}

// contentBrowserCursorDownCmd returns a command to move cursor down and load fields for the next node.
func contentBrowserCursorDownCmd(m Model) tea.Cmd {
	newCursor := m.Cursor + 1
	node := m.Root.NodeAtIndex(newCursor)
	if node != nil && node.Instance != nil {
		return tea.Batch(CursorDownCmd(), LoadContentFieldsCmd(m.Config, node.Instance.ContentDataID, node.Instance.DatatypeID))
	}
	return CursorDownCmd()
}

// contentFieldEdit opens a single-field edit dialog for the currently selected field.
func (m Model) contentFieldEdit() (Model, tea.Cmd) {
	if len(m.SelectedContentFields) == 0 || m.FieldCursor >= len(m.SelectedContentFields) {
		return m, ShowDialog("Info", "No field selected", false)
	}
	cf := m.SelectedContentFields[m.FieldCursor]
	if cf.ContentFieldID.IsZero() {
		return m, ShowDialog("Info", "Field has no value yet. Use 'n' to add.", false)
	}
	node := m.Root.NodeAtIndex(m.Cursor)
	if node == nil || node.Instance == nil {
		return m, nil
	}
	return m, ShowEditSingleFieldDialogCmd(cf, node.Instance.ContentDataID, m.PageRouteId, node.Instance.DatatypeID)
}

// contentFieldAdd shows a picker for fields not yet populated on the content.
func (m Model) contentFieldAdd() (Model, tea.Cmd) {
	node := m.Root.NodeAtIndex(m.Cursor)
	if node == nil || node.Instance == nil {
		return m, ShowDialog("Error", "No content node selected", false)
	}

	// Find missing fields: fields in datatype but not in content
	existingFieldIDs := make(map[string]bool)
	for _, cf := range m.SelectedContentFields {
		if !cf.ContentFieldID.IsZero() {
			existingFieldIDs[string(cf.FieldID)] = true
		}
	}

	// All fields with empty content value are candidates for "add"
	var missing []ContentFieldDisplay
	for _, cf := range m.SelectedContentFields {
		if cf.ContentFieldID.IsZero() {
			missing = append(missing, cf)
		}
	}

	if len(missing) == 0 {
		return m, ShowDialog("Info", "All fields already populated", false)
	}

	// If only one missing field, add it directly
	if len(missing) == 1 {
		return m, m.HandleAddContentField(
			node.Instance.ContentDataID,
			missing[0].FieldID,
			m.PageRouteId,
			node.Instance.DatatypeID,
		)
	}

	// Multiple missing fields - show picker dialog
	options := make([]huh.Option[string], 0, len(missing))
	for _, mf := range missing {
		options = append(options, huh.NewOption(mf.Label, string(mf.FieldID)))
	}
	return m, ShowAddContentFieldDialogCmd(options, node.Instance.ContentDataID, m.PageRouteId, node.Instance.DatatypeID)
}

// contentFieldDelete shows delete confirmation for the selected content field.
func (m Model) contentFieldDelete() (Model, tea.Cmd) {
	if len(m.SelectedContentFields) == 0 || m.FieldCursor >= len(m.SelectedContentFields) {
		return m, ShowDialog("Info", "No field selected", false)
	}
	cf := m.SelectedContentFields[m.FieldCursor]
	if cf.ContentFieldID.IsZero() {
		return m, ShowDialog("Info", "Field has no value to delete", false)
	}
	node := m.Root.NodeAtIndex(m.Cursor)
	if node == nil || node.Instance == nil {
		return m, nil
	}
	return m, ShowDeleteContentFieldDialogCmd(cf, node.Instance.ContentDataID, m.PageRouteId, node.Instance.DatatypeID)
}

// contentFieldReorderUp moves the current field up one position in the field list.
func (m Model) contentFieldReorderUp() (Model, tea.Cmd) {
	if m.FieldCursor <= 0 || len(m.SelectedContentFields) < 2 {
		return m, nil
	}
	node := m.Root.NodeAtIndex(m.Cursor)
	if node == nil || node.Instance == nil {
		return m, nil
	}
	current := m.SelectedContentFields[m.FieldCursor]
	prev := m.SelectedContentFields[m.FieldCursor-1]
	// We need sort_order from the junction records. Since the fields are in sort_order,
	// use index as fallback sort order if DatatypeFieldID is available
	return m, m.HandleReorderField(
		current.DatatypeFieldID, prev.DatatypeFieldID,
		int64(m.FieldCursor), int64(m.FieldCursor-1),
		node.Instance.DatatypeID, node.Instance.ContentDataID, m.PageRouteId, "up",
	)
}

// contentFieldReorderDown moves the current field down one position in the field list.
func (m Model) contentFieldReorderDown() (Model, tea.Cmd) {
	if m.FieldCursor >= len(m.SelectedContentFields)-1 || len(m.SelectedContentFields) < 2 {
		return m, nil
	}
	node := m.Root.NodeAtIndex(m.Cursor)
	if node == nil || node.Instance == nil {
		return m, nil
	}
	current := m.SelectedContentFields[m.FieldCursor]
	next := m.SelectedContentFields[m.FieldCursor+1]
	return m, m.HandleReorderField(
		current.DatatypeFieldID, next.DatatypeFieldID,
		int64(m.FieldCursor), int64(m.FieldCursor+1),
		node.Instance.DatatypeID, node.Instance.ContentDataID, m.PageRouteId, "down",
	)
}

// MediaControls handles keyboard navigation and actions for the media library page.
func (m Model) MediaControls(msg tea.Msg) (Model, tea.Cmd) {
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
			if m.Cursor < len(m.MediaList)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionNew) {
			fp := filepicker.New()
			fp.AllowedTypes = []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}
			fp.CurrentDirectory, _ = os.UserHomeDir()
			fp.Height = m.Height - 4
			m.FilePicker = fp
			m.FilePickerActive = true
			m.FilePickerPurpose = FILEPICKER_MEDIA
			return m, m.FilePicker.Init()
		}
		if km.Matches(key, config.ActionDelete) {
			if len(m.MediaList) > 0 && m.Cursor < len(m.MediaList) {
				media := m.MediaList[m.Cursor]
				label := media.MediaID.String()
				if media.DisplayName.Valid && media.DisplayName.String != "" {
					label = media.DisplayName.String
				} else if media.Name.Valid && media.Name.String != "" {
					label = media.Name.String
				}
				return m, ShowDeleteMediaDialogCmd(media.MediaID, label)
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

// UsersAdminControls handles keyboard navigation and actions for the users admin page.
func (m Model) UsersAdminControls(msg tea.Msg) (Model, tea.Cmd) {
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
			if m.Cursor < len(m.UsersList)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionNew) {
			return m, ShowCreateUserDialogCmd()
		}
		if km.Matches(key, config.ActionEdit) {
			if len(m.UsersList) > 0 && m.Cursor < len(m.UsersList) {
				return m, ShowEditUserDialogCmd(m.UsersList[m.Cursor])
			}
		}
		if km.Matches(key, config.ActionDelete) {
			if len(m.UsersList) > 0 && m.Cursor < len(m.UsersList) {
				user := m.UsersList[m.Cursor]
				return m, ShowDeleteUserDialogCmd(user.UserID, user.Username)
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

// BasicContentControls handles keyboard navigation for content-related pages.
func (m Model) BasicContentControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
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
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.PageMenu) {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			page := m.PageMenu[m.Cursor]
			switch page.Index {
			default:
				return m, nil
			}
		}
	}
	return m, nil
}


// BasicDynamicControls handles keyboard navigation for dynamic pages.
func (m Model) BasicDynamicControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
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
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.PageMenu)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			if len(m.DatatypeMenu) > 0 {
				return m, tea.Batch(
					NavigateToPageCmd(m.Pages[DYNAMICPAGE]),
				)
			}
		}
	}
	return m, nil
}
// SelectTable handles keyboard navigation for selecting a database table.
func (m Model) SelectTable(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.Tables)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			cmds = append(cmds, SelectTableCmd(m.Tables[m.Cursor]))
		}
	}
	return m, tea.Batch(cmds...)
}

// TableNavigationControls handles keyboard navigation for browsing table rows.
func (m Model) TableNavigationControls(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()
		handled := false

		if km.Matches(key, config.ActionQuit) || km.Matches(key, config.ActionDismiss) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
			handled = true
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(m.TableState.Rows)-1 {
				return m, CursorDownCmd()
			}
			handled = true
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			// "h", "shift+tab", "backspace" go back in history
			if key != "left" {
				if len(m.History) > 0 {
					return m, HistoryPopCmd()
				}
				handled = true
			}
		}
		if km.Matches(key, config.ActionPagePrev) {
			if m.PageMod > 0 {
				return m, PageModPreviousCmd()
			}
			handled = true
		}
		if km.Matches(key, config.ActionPageNext) {
			if m.PageMod < (len(m.TableState.Rows)-1)/m.MaxRows {
				return m, PageModNextCmd()
			}
			handled = true
		}
		// Select action (enter, l)
		if key == "enter" || key == "l" {
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			if recordIndex < len(m.TableState.Rows) {
				cmds = append(cmds, CursorSetCmd(recordIndex))

				switch m.Page.Index {
				case READPAGE:
					cmds = append(cmds, NavigateToPageCmd(m.PageMap[READSINGLEPAGE]))
				case UPDATEPAGE:
					cmds = append(cmds, ShowDatabaseUpdateDialogCmd(db.DBTable(m.TableState.Table), m.TableState.Rows[recordIndex][0]))
				case DELETEPAGE:
					cmds = append(cmds, ShowDialogCmd("Confirm Delete",
						"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE))
				}
			}
			handled = true
		}

		if !handled {
			cmds = append(cmds, m.UpdateMaxCursorCmd())
			cmds = append(cmds, PaginationUpdateCmd())
		}
	}
	return m, tea.Batch(cmds...)
}

// UpdateDatabaseDelete handles keyboard events for confirming a record deletion.
func (m Model) UpdateDatabaseDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()
		handled := false

		if km.Matches(key, config.ActionQuit) {
			return m, tea.Quit
		}
		if key == "enter" || key == "l" {
			err := m.DatabaseDelete(m.Config, db.StringDBTable(m.TableState.Table))
			if err != nil {
				return m, nil
			}
			if m.Cursor > 0 {
				m.Cursor--
			}
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, nil))
			handled = true
		}
		if !handled {
			var scmd tea.Cmd
			m.Spinner, scmd = m.Spinner.Update(msg)
			cmds = append(cmds, scmd)
		}
	}
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
	cmds = append(cmds, pcmd)
	return m, tea.Batch(cmds...)
}
// DefineDatatypeControls handles keyboard events for the datatype definition form.
func (m Model) DefineDatatypeControls(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	cmds = append(cmds, FocusSetCmd(FORMFOCUS))

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			m.Logger.Finfo("Tables Fetch ")
		}
	}()

	// Update form with the message
	form, cmd := m.FormState.Form.Update(msg)
	if _, ok := form.(*huh.Form); ok {
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.FormState.Form.State == huh.StateAborted {
		datatypesPage := m.PageMap[DATATYPES]
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, FormCompletedCmd(&datatypesPage))
	}

	if m.FormState.Form.State == huh.StateCompleted {
		datatypesPage := m.PageMap[DATATYPES]
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, LoadingStartCmd())
		cmds = append(cmds, AllDatatypesFetchCmd())
		cmds = append(cmds, FormCompletedCmd(&datatypesPage))
	}

	return m, tea.Batch(cmds...)
}

// DevelopmentInterface handles keyboard events for the development/debug page.
func DevelopmentInterface(m Model, message tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionPrevPanel) {
			if len(m.History) > 0 {
				entry := *m.PopHistory()
				m.Page = entry.Page
				m.Cursor = entry.Cursor
				return m, nil
			}
		}
		if km.Matches(key, config.ActionDelete) {
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

// ActionsControls handles keyboard navigation for the actions menu page.
func (m Model) ActionsControls(msg tea.Msg) (Model, tea.Cmd) {
	actions := ActionsMenu()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := m.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) {
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
			return m, tea.Quit
		}
		if km.Matches(key, config.ActionUp) {
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		}
		if km.Matches(key, config.ActionDown) {
			if m.Cursor < len(actions)-1 {
				return m, CursorDownCmd()
			}
		}
		if km.Matches(key, config.ActionSelect) {
			if m.Cursor >= len(actions) {
				return m, nil
			}
			action := actions[m.Cursor]
			if action.Destructive {
				return m, func() tea.Msg {
					return ActionConfirmMsg{ActionIndex: m.Cursor}
				}
			}
			return m, tea.Batch(
				LoadingStartCmd(),
				RunActionCmd(ActionParams{
					Config:         m.Config,
					UserID:         m.UserID,
					SSHFingerprint: m.SSHFingerprint,
					SSHKeyType:     m.SSHKeyType,
					SSHPublicKey:   m.SSHPublicKey,
				}, m.Cursor),
			)
		}
	}
	return m, nil
}

// ConfigControls handles viewport scrolling for the configuration display page.
func (m Model) ConfigControls(msg tea.Msg) (Model, tea.Cmd) {
	newModel := m

	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := newModel.Config.KeyBindings
		key := msg.String()

		if km.Matches(key, config.ActionQuit) {
			return newModel, tea.Quit
		}
		if km.Matches(key, config.ActionBack) || km.Matches(key, config.ActionDismiss) {
			if len(newModel.History) > 0 {
				return newModel, HistoryPopCmd()
			}
			return newModel, tea.Quit
		}
	}

	// Forward all other events to the viewport for scrolling
	var cmd tea.Cmd
	newModel.Viewport, cmd = newModel.Viewport.Update(msg)
	return newModel, cmd
}
