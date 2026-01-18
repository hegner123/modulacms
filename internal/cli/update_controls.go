package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
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
	case CONTENT:
		return m.ContentBrowserControls(msg)
	case USERSADMIN:
		return m.BasicCMSControls(msg)
	case MEDIA:
		return m.BasicCMSControls(msg)

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
			page := m.PageMenu[m.Cursor]
			switch page.Index {
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

// ContentBrowserControls handles keyboard navigation for the content browser
func (m Model) ContentBrowserControls(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Exit / Back
		case "q", "esc", "h", "left", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit

		// Navigation
		case "up", "k":
			if m.Cursor > 0 {
				return m, CursorUpCmd()
			}
		case "down", "j":
			// Count visible nodes
			maxCursor := m.countVisibleNodes()
			if m.Cursor < maxCursor-1 {
				return m, CursorDownCmd()
			}

		// Expand/Collapse
		case "enter", "space":
			node := m.getSelectedNodeFromModel()
			if node != nil && node.FirstChild != nil {
				node.Expand = !node.Expand
				return m, nil
			}

		// Actions
		case "e":
			// Edit selected content
			return m, NavigateToPageCmd(m.PageMap[EDITCONTENT])
		case "n":
			// Create new content - build form based on selected datatype
			node := m.getSelectedNodeFromModel()
			utility.DefaultLogger.Finfo(fmt.Sprintf("'n' key pressed, node: %v", node != nil))
			if node != nil {
				utility.DefaultLogger.Finfo(fmt.Sprintf("Building form for datatype %d (label: %s)", node.Datatype.DatatypeID, node.Datatype.Label))
				// Use the datatype from the selected node
				return m, BuildContentFormCmd(node.Datatype.DatatypeID, m.PageRouteId)
			}
			// Fallback: if no node selected, we can't create content
			utility.DefaultLogger.Finfo("No node selected")
			return m, ShowDialog("Error", "Please select a content type first", false)
		case "d":
			// Delete content (with confirmation)
			return m, ShowDialog("Confirm Delete", "Delete this content? This cannot be undone.", true)

		// Search
		case "/":
			// TODO: Implement search
			return m, nil

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

func (m Model) ConfigControls(msg tea.Msg) (Model, tea.Cmd) {
	newModel:=m
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the viewport
	newModel.Viewport, cmd = newModel.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return newModel, tea.Batch(cmds...)

}
