package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
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
		return m.FormControls(msg)
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
				return m, tea.Batch(
					CmsDefineDatatypeLoadCmd(),
				)
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
		}
	}
	return m, nil
}

// ContentBrowserControls handles keyboard navigation for the content browser.
// The Content page has a multi-step flow:
//   - Left panel: ROOT datatypes list
//   - Center panel: routes for selected ROOT type (or content tree if route active)
//   - Right panel: details/actions
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc", "h", "left", "backspace":
			// Step back through content flow states
			if !m.PageRouteId.IsZero() {
				// Clear route, go back to route list
				m.PageRouteId = types.RouteID("")
				m.Root = TreeRoot{}
				m.Cursor = 0
				return m, nil
			}
			if !m.SelectedDatatype.IsZero() {
				// Clear datatype, go back to ROOT type list
				m.SelectedDatatype = types.DatatypeID("")
				m.Routes = nil
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
				// Navigating ROOT types or routes list
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
				if m.SelectedDatatype.IsZero() {
					// Navigating ROOT types
					if m.Cursor < len(m.RootDatatypes)-1 {
						return m, CursorDownCmd()
					}
				} else {
					// Navigating routes list
					if m.Cursor < len(m.Routes)-1 {
						return m, CursorDownCmd()
					}
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
				if m.SelectedDatatype.IsZero() {
					// Select a ROOT datatype -> fetch routes for it
					if len(m.RootDatatypes) > 0 && m.Cursor < len(m.RootDatatypes) {
						dt := m.RootDatatypes[m.Cursor]
						m.SelectedDatatype = dt.DatatypeID
						m.Cursor = 0
						return m, tea.Batch(
							SelectedDatatypeSetCmd(dt.DatatypeID),
							RoutesByDatatypeFetchCmd(dt.DatatypeID),
						)
					}
				} else {
					// Select a route -> set PageRouteId and load content tree
					if len(m.Routes) > 0 && m.Cursor < len(m.Routes) {
						route := m.Routes[m.Cursor]
						m.PageRouteId = route.RouteID
						m.Cursor = 0
						return m, tea.Batch(
							LogMessageCmd(fmt.Sprintf("Route selected: %s (%s)", route.Title, route.RouteID)),
							ReloadContentTreeCmd(m.Config, route.RouteID),
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

		// Actions (only when browsing content tree)
		case "e":
			if !m.PageRouteId.IsZero() {
				return m, NavigateToPageCmd(m.PageMap[EDITCONTENT])
			}
		case "n":
			if !m.PageRouteId.IsZero() {
				node := m.getSelectedNodeFromModel()
				utility.DefaultLogger.Finfo(fmt.Sprintf("'n' key pressed, node: %v", node != nil))
				if node != nil {
					utility.DefaultLogger.Finfo(fmt.Sprintf("Building form for datatype %s (label: %s)", node.Datatype.DatatypeID, node.Datatype.Label))
					return m, m.BuildContentFieldsForm(node.Datatype.DatatypeID, m.PageRouteId)
				}
				utility.DefaultLogger.Finfo("No node selected")
				return m, ShowDialog("Error", "Please select a content type first", false)
			}
		case "d":
			if !m.PageRouteId.IsZero() {
				return m, ShowDialog("Confirm Delete", "Delete this content? This cannot be undone.", true)
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
			utility.DefaultLogger.Finfo("Tables Fetch ")
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
			cmd := FetchTableHeadersRowsCmd(*m.Config, m.TableState.Table, nil)
			cmds = append(cmds, cmd)
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
			utility.DefaultLogger.Finfo("Tables Fetch ")
		}
	}()

	// Update form with the message
	form, cmd := m.FormState.Form.Update(msg)
	if _, ok := form.(*huh.Form); ok {
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.FormState.Form.State == huh.StateAborted {
		cmds = append(cmds, FormCancelCmd())
	}

	if m.FormState.Form.State == huh.StateCompleted {
		utility.DefaultLogger.Finfo("Tables Fetch ")
		// TODO: Implement form completion handling with proper messages
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
				RunActionCmd(m.Config, m.Cursor),
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
