package cli

import (
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
	case DATABASEPAGE:
		return m.SelectTable(msg)
	case TABLEPAGE:
		return m.BasicControls(msg)
	case CMSPAGE:
		return m.BasicCMSControls(msg)
	case DYNAMICPAGE:
		return m.BasicDynamicControls(msg)
	case CREATEPAGE:
		return m.FormControls(msg)
	case READPAGE:
		return m.TableNavigationControls(msg)
	case UPDATEPAGE:
		return m.TableNavigationControls(msg)
	case DELETEPAGE:
		return m.TableNavigationControls(msg)
	case DEVELOPMENT:
		return DevelopmentInterface(m, msg)
	case DATATYPE:
		return m.DefineDatatypeControls(msg)
	case CONFIGPAGE:
		return m.ConfigControls(msg)

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
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l":
			// Only proceed if we have menu items
			if len(m.PageMenu) > 0 {
				return m, NavigateToPageCmd(*m.PageMenu[m.Cursor])
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
			if m.Cursor < len(m.DatatypeMenu)-1 {
				return m, CursorDownCmd()
			}
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l":
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
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l":
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
		case "h", "shift+tab", "backspace":
			if len(m.History) > 0 {
				return m, HistoryPopCmd()
			}
		case "enter", "l":
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
	if newModel.Form == nil {
		return newModel, nil
	}

	form, cmd := newModel.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		newModel.Form = f
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Handle form state changes
	if newModel.Form.State == huh.StateAborted {
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, HistoryPopCmd())
	}

	if newModel.Form.State == huh.StateCompleted {
		cmds = append(cmds, FocusSetCmd(PAGEFOCUS))
		cmds = append(cmds, FormSubmitCmd())
		cmds = append(cmds, HistoryPopCmd())
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
			if m.Cursor < len(m.Rows)-1 {
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
			if m.PageMod < (len(m.Rows)-1)/m.MaxRows {
				return m, PageModNextCmd()
			}

		//Action
		case "enter", "l":
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			if recordIndex < len(m.Rows) {
				cmds = append(cmds, CursorSetCmd(recordIndex))

				// Handle different actions based on current page
				switch m.Page.Index {
				case READPAGE:
					// Navigate to single record view
					cmds = append(cmds, NavigateToPageCmd(m.Pages[READSINGLEPAGE]))
				case UPDATEPAGE:
					// Navigate to update form with pre-populated values
					cmds = append(cmds, NavigateToPageCmd(m.Pages[UPDATEFORMPAGE]))
				case DELETEPAGE:
					// Show confirmation dialog
					cmds = append(cmds, ShowDialogCmd("Confirm Delete",
						"Are you sure you want to delete this record? This action cannot be undone.", true, DIALOGDELETE))
				}
			}

		default:
			cmds = append(cmds, m.UpdateMaxCursor())
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
			if m.PageMod < len(m.Rows)/m.MaxRows {
				m.PageMod++
			}

		//Action
		case "enter", "l":
			rows = m.Rows
			m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})
			recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
			// Only update if the calculated index is valid
			if recordIndex < len(m.Rows) {
				m.Cursor = recordIndex
			}
			m.Row = &rows[recordIndex]
			m.Cursor = 0
			m.Page = m.Pages[UPDATEFORMPAGE]
			m.PageMenu = m.Page.Children

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
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.Form.State == huh.StateAborted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[UPDATEPAGE]
		cmd := m.DatabaseUpdate(m.Config, db.DBTable(m.Table))
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
			err := m.DatabaseDelete(m.Config, db.StringDBTable(m.Table))
			if err != nil {
				return m, nil
			}
			if m.Cursor > 0 {
				m.Cursor--
			}
			cmd := FetchTableHeadersRowsCmd(*m.Config, m.Table)
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
	form, cmd := m.Form.Update(msg)
	if _, ok := form.(*huh.Form); ok {
		cmds = append(cmds, cmd)
	}

	// Handle form state changes
	if m.Form.State == huh.StateAborted {
		cmds = append(cmds, FormCancelCmd())
	}

	if m.Form.State == huh.StateCompleted {
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
				m.PageMenu = m.Page.Children
				return m, nil
			}
		case "d":
			return m, nil
		}

	}

	return m, tea.Batch(cmds...)

}

func (m Model) ConfigControls(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// Handle keyboard and mouse events in the viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)

}
