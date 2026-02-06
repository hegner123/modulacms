package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tui"
)

type ControlUpdate struct{}

func NewControlUpdate() tea.Cmd {
	return func() tea.Msg {
		return ControlUpdate{}
	}
}

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
		return m.BasicControls(msg)
	case DYNAMICPAGE:
		return m.BasicDynamicControls(msg)
	case CREATEPAGE:
		return m.FormControls(msg)
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
		return m.BasicCMSControls(msg)
	case MEDIA:
		return m.MediaControls(msg)

	}
	return m, nil
}

func (m Model) BasicControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept form navigation keys when form has focus
		switch msg.String() {
		//Exit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.PageMenu)-1 {
				return m, CursorDownCmd()
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			// Only proceed if we have menu items
			if len(m.PageMenu) > 0 {
				return m, NavigateToPageCmd(m.PageMenu[m.Cursor])
			}
		}
	}
	return m, nil
}

func (m Model) BasicCMSControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept form navigation keys when form has focus
		switch msg.String() {
		//Exit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		case "shift+tab":
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil

		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.PageMenu) {
				return m, CursorDownCmd()
			}
		case "h", "left", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			// Only proceed if we have menu items
			page := m.PageMenu[m.Cursor]
			switch page.Index {
			case ROUTES:
				return m, NavigateToPageCmd(m.PageMap[ROUTES])
			case DATATYPES:
				return m, NavigateToPageCmd(m.PageMap[DATATYPES])
			case FIELDS:
				return m, tea.Batch()
			case CONTENT:
				// Navigate to content browser
				return m, NavigateToPageCmd(m.PageMap[CONTENT])
			case MEDIA:
				// Navigate to media page
				return m, NavigateToPageCmd(m.PageMap[MEDIA])
			case USERSADMIN:
				// Navigate to users admin page
				return m, NavigateToPageCmd(m.PageMap[USERSADMIN])
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
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		case "shift+tab":
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.Routes)-1 {
				return m, CursorDownCmd()
			}
		case "h", "left", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
				route := m.Routes[m.Cursor]
				m.PageRouteId = route.RouteID
				return m, LogMessageCmd(fmt.Sprintf("Route selected: %s (%s)", route.Title, route.RouteID))
			}
		case "n":
			// Create new route
			return m, ShowRouteFormDialogCmd(FORMDIALOGCREATEROUTE, "New Route")
		case "e":
			// Edit selected route
			if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
				return m, ShowEditRouteDialogCmd(m.Routes[m.Cursor])
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
func (m Model) DatatypesControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Escape goes back to tree panel if not already there
			if m.PanelFocus != tui.TreePanel {
				m.PanelFocus = tui.TreePanel
				return m, nil
			}
			return m, tea.Quit
		case "tab":
			m.PanelFocus = (m.PanelFocus + 1) % 3
			// Reset field cursor when entering content panel
			if m.PanelFocus == tui.ContentPanel {
				m.FieldCursor = 0
			}
			return m, nil
		case "shift+tab":
			m.PanelFocus = (m.PanelFocus + 2) % 3
			// Reset field cursor when entering content panel
			if m.PanelFocus == tui.ContentPanel {
				m.FieldCursor = 0
			}
			return m, nil
		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		case "up", "k":
			return m.datatypesControlsUp()
		case "down", "j":
			return m.datatypesControlsDown()
		case "h", "left", "backspace":
			// If in center or right panel, move left to tree panel
			if m.PanelFocus != tui.TreePanel {
				m.PanelFocus = tui.TreePanel
				return m, nil
			}
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "l", "right":
			// Move right to next panel
			if m.PanelFocus == tui.TreePanel {
				m.PanelFocus = tui.ContentPanel
				m.FieldCursor = 0
				return m, nil
			}
			if m.PanelFocus == tui.ContentPanel {
				m.PanelFocus = tui.RoutePanel
				return m, nil
			}
		case "n":
			return m.datatypesControlsNew()
		case "e":
			return m.datatypesControlsEdit()
		case "enter":
			return m.datatypesControlsSelect()
		}
	}
	return m, nil
}

// datatypesControlsUp handles up navigation based on panel focus
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

// datatypesControlsDown handles down navigation based on panel focus
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

// datatypesControlsNew handles 'n' key based on panel focus
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

// datatypesControlsEdit handles 'e' key based on panel focus
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

// datatypesControlsSelect handles enter key based on panel focus
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
		switch msg.String() {
		// Panel focus cycling
		case "tab":
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		case "shift+tab":
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil

		// Exit / Back
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Show quit confirmation dialog
			return m, ShowQuitConfirmDialogCmd()
		case "esc", "h", "left", "backspace":
			// Step back through content flow states
			if !m.PageRouteId.IsZero() {
				// Clear route, go back to content list
				m.PageRouteId = types.RouteID("")
				m.Root = TreeRoot{}
				m.Cursor = 0
				return m, nil
			}
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
			return m, tea.Quit

		// Navigation
		case "up", "k":
			if m.PageRouteId.IsZero() {
				// Navigating content summary list
				if m.Cursor > 0 {
					return m, CursorUpCmd()
				}
			} else {
				// Navigating content tree
				if m.Cursor > 0 {
					return m, CursorUpCmd()
				}
			}
		case "down", "j":
			if m.PageRouteId.IsZero() {
				// Navigating content summary list
				if m.Cursor < len(m.RootContentSummary)-1 {
					return m, CursorDownCmd()
				}
			} else {
				// Navigating content tree
				maxCursor := m.countVisibleNodes()
				if m.Cursor < maxCursor-1 {
					return m, CursorDownCmd()
				}
			}

		// Selection
		case "enter", "l", "right":
			if m.PageRouteId.IsZero() {
				// Select a content summary -> go to content tree
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
				// Content tree: expand/collapse
				node := m.getSelectedNodeFromModel()
				if node != nil && node.FirstChild != nil {
					node.Expand = !node.Expand
					return m, nil
				}
			}

		// Actions
		case "e":
			if !m.PageRouteId.IsZero() {
				// Edit the selected node - show content fields dialog with existing values
				node := m.getSelectedNodeFromModel()
				if node != nil && node.Instance != nil {
					m.Logger.Finfo(fmt.Sprintf("'e' key pressed, editing node %s with datatype %s", node.Instance.ContentDataID, node.Datatype.Label))
					// Fetch existing content fields and show edit dialog
					return m, FetchContentForEditCmd(
						node.Instance.ContentDataID,
						node.Datatype.DatatypeID,
						m.PageRouteId,
						fmt.Sprintf("Edit: %s", node.Datatype.Label),
					)
				}
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		case "n":
			if m.PageRouteId.IsZero() {
				// Create a new route with content using ROOT datatype from selected content
				if len(m.RootContentSummary) > 0 && m.Cursor < len(m.RootContentSummary) {
					content := m.RootContentSummary[m.Cursor]
					if content.DatatypeID.Valid {
						return m, ShowCreateRouteWithContentDialogCmd(string(content.DatatypeID.ID))
					}
				} else if len(m.RootDatatypes) > 0 {
					// Fallback: use first ROOT datatype if no content selected
					return m, ShowCreateRouteWithContentDialogCmd(string(m.RootDatatypes[0].DatatypeID))
				}
			} else {
				// When viewing content tree, show dialog to select child datatype
				node := m.getSelectedNodeFromModel()
				m.Logger.Finfo(fmt.Sprintf("'n' key pressed, node: %v", node != nil))
				if node != nil {
					m.Logger.Finfo(fmt.Sprintf("Showing child datatype picker for parent %s", node.Datatype.Label))
					return m, ShowChildDatatypeDialogCmd(node.Datatype.DatatypeID, m.PageRouteId)
				}
				m.Logger.Finfo("No node selected")
				return m, ShowDialog("Error", "Please select a content node first", false)
			}
		case "d":
			if !m.PageRouteId.IsZero() {
				node := m.getSelectedNodeFromModel()
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

		// Title font change
		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}
	}
	return m, nil
}

// countVisibleNodes counts the number of visible nodes in the tree
func (m Model) countVisibleNodes() int {
	if m.Root.Root == nil {
		return 0
	}
	count := 0
	m.countNodesRecursive(m.Root.Root, &count)
	return count
}

// countNodesRecursive recursively counts visible nodes
func (m Model) countNodesRecursive(node *TreeNode, count *int) {
	if node == nil {
		return
	}
	*count++

	if node.Expand && node.FirstChild != nil {
		m.countNodesRecursive(node.FirstChild, count)
	}

	if node.NextSibling != nil {
		m.countNodesRecursive(node.NextSibling, count)
	}
}

// getSelectedNodeFromModel returns the currently selected node
func (m Model) getSelectedNodeFromModel() *TreeNode {
	if m.Root.Root == nil {
		return nil
	}
	currentIndex := 0
	return m.findNodeAtIndex(m.Root.Root, m.Cursor, &currentIndex)
}

// findNodeAtIndex finds the node at the given index
func (m Model) findNodeAtIndex(node *TreeNode, targetIndex int, currentIndex *int) *TreeNode {
	if node == nil {
		return nil
	}

	if *currentIndex == targetIndex {
		return node
	}
	*currentIndex++

	if node.Expand && node.FirstChild != nil {
		if result := m.findNodeAtIndex(node.FirstChild, targetIndex, currentIndex); result != nil {
			return result
		}
	}

	if node.NextSibling != nil {
		return m.findNodeAtIndex(node.NextSibling, targetIndex, currentIndex)
	}

	return nil
}

// MediaControls handles keyboard navigation for the media library page.
func (m Model) MediaControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.PanelFocus = (m.PanelFocus + 1) % 3
			return m, nil
		case "shift+tab":
			m.PanelFocus = (m.PanelFocus + 2) % 3
			return m, nil
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.MediaList)-1 {
				return m, CursorDownCmd()
			}
		case "h", "left", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		}
	}
	return m, nil
}

func (m Model) BasicContentControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		//Exit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.PageMenu) {
				return m, CursorDownCmd()
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			// Only proceed if we have menu items
			// Use datatypes as menu items
			// Next page has AddNew and a list of existing content
			page := m.PageMenu[m.Cursor]
			switch page.Index {
			default:
				return m, nil
			}
		}
	}
	return m, nil
}


func (m Model) BasicDynamicControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept form navigation keys when form has focus
		switch msg.String() {
		//Exit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "shift+left":
			if m.TitleFont > 0 {
				return m, TitleFontPreviousCmd()
			}
		case "shift+right":
			if m.TitleFont < len(m.Titles)-1 {
				return m, TitleFontNextCmd()
			}
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.PageMenu)-1 {
				return m, CursorDownCmd()
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			// Only proceed if we have menu items
			if len(m.DatatypeMenu) > 0 {
				return m, tea.Batch(
					NavigateToPageCmd(m.Pages[DYNAMICPAGE]),
				)
			}
		}
	}
	return m, nil
}
func (m Model) SelectTable(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.Tables)-1 {
				return m, CursorDownCmd()
			}
		case "h", "left", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l", "right":
			cmds = append(cmds, SelectTableCmd(m.Tables[m.Cursor]))
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) FormControls(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	newModel := m
	newModel.Focus = FORMFOCUS

	// Ensure form exists before updating
	if newModel.FormState.Form == nil {
		return newModel, nil
	}

	form, cmd := newModel.FormState.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		newModel.FormState.Form = f
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Handle form state changes
	if newModel.FormState.Form.State == huh.StateAborted {
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, FormCompletedCmd(nil)) // Will try history pop, then home
	}

	if newModel.FormState.Form.State == huh.StateCompleted {
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, FormSubmitCmd())
		// Note: Don't navigate here - let specific form completion messages
		// (like ContentCreatedMsg) handle navigation after async operations complete
	}

	return newModel, tea.Batch(cmds...)
}

func (m Model) TableNavigationControls(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(m.TableState.Rows)-1 {
				return m, CursorDownCmd()
			}
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "left":
			if m.PageMod > 0 {
				return m, PageModPreviousCmd()
			}
		case "right":
			if m.PageMod < (len(m.TableState.Rows)-1)/m.MaxRows {
				return m, PageModNextCmd()
			}

		//Action
		case "enter", "l":
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			if recordIndex < len(m.TableState.Rows) {
				cmds = append(cmds, CursorSetCmd(recordIndex))

				// Handle different actions based on current page
				switch m.Page.Index {
				case READPAGE:
					// Navigate to single record view
					cmds = append(cmds, NavigateToPageCmd(m.PageMap[READSINGLEPAGE]))
				case UPDATEPAGE:
					// Navigate to update form with pre-populated values
					cmds = append(cmds, NavigateToPageCmd(m.PageMap[UPDATEFORMPAGE]))
				case DELETEPAGE:
					// Show confirmation dialog
					cmds = append(cmds, ShowDialogCmd("Confirm Delete",
						"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE))
				}
			}

		default:
			cmds = append(cmds, m.UpdateMaxCursorCmd())
			cmds = append(cmds, PaginationUpdateCmd())
		}
	}
	return m, tea.Batch(cmds...)

}

func (m Model) UpdateDatabaseUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var rows [][]string
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		//Exit
		case "left":
			if m.PageMod > 0 {
				m.PageMod--
			}
		case "right":
			if m.PageMod < len(m.TableState.Rows)/m.MaxRows {
				m.PageMod++
			}

		//Action
		case "enter", "l":
			rows = m.TableState.Rows
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			// Only update if the calculated index is valid
			if recordIndex < len(m.TableState.Rows) {
				m.Cursor = recordIndex
			}
			m.TableState.Row = &rows[recordIndex]
			m.Cursor = 0
			m.Page = m.Pages[UPDATEFORMPAGE]

		}
	}
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
	cmds = append(cmds, pcmd)
	return m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseFormUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.Focus = FORMFOCUS

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
	if f, ok := form.(*huh.Form); ok {
		m.FormState.Form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.FormState.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
	}

	if m.FormState.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
		cmd := m.DatabaseUpdate(m.Config, db.DBTable(m.TableState.Table))
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.Spinner, scmd = m.Spinner.Update(msg)
	cmds = append(cmds, scmd)

	return m, tea.Batch(cmds...)
}

func (m Model) UpdateDatabaseDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter", "l":
			err := m.DatabaseDelete(m.Config, db.StringDBTable(m.TableState.Table))
			if err != nil {
				return m, nil
			}
			if m.Cursor > 0 {
				m.Cursor--
			}
			cmds = append(cmds, LoadingStartCmd())
			cmds = append(cmds, FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, nil))
		default:
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

func DevelopmentInterface(m Model, message tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		//Exit
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				entry := *m.PopHistory()
				m.Page = entry.Page
				m.Cursor = entry.Cursor
				return m, nil
			}
		case "d":
			return m, nil
		}

	}

	return m, tea.Batch(cmds...)

}

func (m Model) ActionsControls(msg tea.Msg) (Model, tea.Cmd) {
	actions := ActionsMenu()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "h", "left", "backspace", "esc":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			if m.Cursor < len(actions)-1 {
				return m, CursorDownCmd()
			}
		case "enter", "l", "right":
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

func (m Model) ConfigControls(msg tea.Msg) (Model, tea.Cmd) {
	newModel := m

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return newModel, tea.Quit
		case "h", "backspace", "esc":
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
