package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/cli/cms"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/utility"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var keyMsg tea.KeyMsg
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
		headerHeight := lipgloss.Height(m.headerView() + RenderTitle(m.Titles[m.TitleFont]) + RenderHeading(m.Header))
		footerHeight := lipgloss.Height(m.footerView() + RenderFooter(m.Footer))
		verticalMarginHeight := headerHeight + footerHeight

		if !m.Ready {
			m.Viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.Viewport.YPosition = headerHeight
			m.Ready = true
		} else {
			m.Viewport.YPosition = headerHeight
			m.Viewport.Width = msg.Width - 4
			m.Viewport.Height = msg.Height - verticalMarginHeight - 10
		}
	case LogMsg:
		utility.DefaultLogger.Finfo(msg.Message)
		return m, nil
	case ClearScreen:
		return m, tea.ClearScreen
	case LoadingTrue:
		newModel := m
		newModel.Loading = true
		return newModel, nil
	case LoadingFalse:
		newModel := m
		newModel.Loading = false
		return newModel, nil
	case CursorUp:
		newModel := m
		newModel.Cursor = m.Cursor - 1
		return newModel, nil
	case CursorDown:
		newModel := m
		newModel.Cursor = m.Cursor + 1
		return newModel, nil
	case CursorReset:
		newModel := m
		newModel.Cursor = 0
		return newModel, nil
	case CursorSet:
		newModel := m
		newModel.Cursor = msg.Index
		return newModel, nil
	case FocusSet:
		newModel := m
		newModel.Focus = msg.Focus
		return newModel, nil
	case TitleFontNext:
		newModel := m
		if newModel.TitleFont < len(m.Titles)-1 {
			newModel.TitleFont++
		}
		return newModel, LogMessage("Title Next Font")
	case TitleFontPrevious:
		newModel := m
		if newModel.TitleFont > 0 {
			newModel.TitleFont--
		}
		return newModel, nil
	case TableSet:
		newModel := m
		newModel.Table = msg.Table
		return newModel, nil
	case TablesSet:
		newModel := m
		newModel.Tables = msg.Tables
		return newModel, nil
	case PageSet:
		newModel := m
		newModel.Page = msg.Page
		return newModel, nil
	case HistoryPush:
		newModel := m
		newModel.History = append(newModel.History, msg.Page)
		return newModel, nil

	case NavigateToPage:
		var cmds []tea.Cmd
		cmds = append(cmds, HistoryPushCmd(PageHistory{Page: m.Page, Cursor: m.Cursor}))
		cmds = append(cmds, CursorResetCmd())
		switch msg.Page.Index {
		case DATABASEPAGE:
			cmds = append(cmds, TablesFetchCmd())
			cmds = append(cmds, PageSetCmd(msg.Page))
		case TABLEPAGE:
			cmds = append(cmds, PageMenuSetCmd(TableMenu))
			cmds = append(cmds, PageSetCmd(msg.Page))
		case CREATEPAGE:
			formCmd := m.BuildCreateDBForm(db.StringDBTable(m.Table))
			m.Form.Init()

			return m, tea.Batch(
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				formCmd,
				FocusSetCmd(FORMFOCUS),
				PageSetCmd(m.Pages[CREATEPAGE]),
				StatusSetCmd(EDITING),
			)
		case UPDATEPAGE:
			return m, tea.Batch(
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[UPDATEPAGE]),
				StatusSetCmd(OK),
			)
		case READPAGE:
			return m, tea.Batch(
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[READPAGE]),
				StatusSetCmd(OK),
			)
		case DELETEPAGE:
			return m, tea.Batch(
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[DELETEPAGE]),
				StatusSetCmd(DELETING),
			)
		case UPDATEFORMPAGE:
			return m, tea.Batch(
				m.BuildUpdateDBForm(db.StringDBTable(m.Table)),
				FetchTableHeadersRowsCmd(*m.Config, m.Table),
				PageSetCmd(m.Pages[UPDATEFORMPAGE]),
				StatusSetCmd(EDITING),
			)
		case READSINGLEPAGE:
			return m, tea.Batch(
				PageSetCmd(m.Pages[READSINGLEPAGE]),
			)
		case CONFIGPAGE:
			form, err := formatJSON(m.Config)
			if err == nil {
				m.Content = form
			}
			m.Viewport.SetContent(m.Content)
			m.Ready = true

			if len(m.PageMenu) > 0 && m.Cursor < len(m.PageMenu) {
				cmds = append(cmds, PageSetCmd(*m.PageMenu[m.Cursor]))
			}

			return m, tea.Batch(cmds...)
		}
		cmds = append(cmds, LogMessage(fmt.Sprintln("cursor:", m.Cursor)))
		cmds = append(cmds, LogMessage(fmt.Sprintln("pages", ViewPageMenus(m))))

		return m, tea.Batch(
			cmds...,
		)

	case HistoryPop:
		newModel := m
		entry := m.PopHistory()
		newModel.PageMenu = m.Page.Children
		return newModel, tea.Batch(
			PageSetCmd(entry.Page),
			CursorSetCmd(entry.Cursor),
		)
	case SelectTable:
		return m, tea.Batch(
			NavigateToPageCmd(m.Pages[TABLEPAGE]),
			TableSetCmd(m.Tables[m.Cursor]),
			PageMenuSetCmd(TableMenu),
		)
	case NavigateToDatabaseCreate:
		return m, tea.Batch(
			NavigateToPageCmd(m.Pages[CREATEPAGE]),
		)

	case FetchHeadersRows:
		c := msg.Config
		t := msg.Table
		dbt := db.StringDBTable(t)
		query := "SELECT * FROM"
		d := db.ConfigDB(c)
		rows, err := d.ExecuteQuery(query, dbt)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}
		defer rows.Close()
		columns, err := rows.Columns()
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}
		listRows, err := db.GenericList(dbt, d)
		if err != nil {
			utility.DefaultLogger.Ferror("", err)
			return m, ErrorSetCmd(err)
		}
		return m, TableHeadersRowsFetchedCmd(columns, listRows)
	case ColumnsSet:
		newModel := m
		newModel.Columns = msg.Columns
		return newModel, nil
	case ColumnTypesSet:
		newModel := m
		newModel.ColumnTypes = msg.ColumnTypes
		return newModel, nil
	case HeadersSet:
		newModel := m
		newModel.Headers = msg.Headers
		return newModel, nil
	case RowsSet:
		newModel := m
		newModel.Rows = msg.Rows
		return newModel, nil
	case CursorMaxSet:
		newModel := m
		newModel.CursorMax = msg.CursorMax
		return newModel, nil
	case PaginatorUpdate:
		newModel := m
		newModel.Paginator.PerPage = msg.PerPage
		newModel.Paginator.SetTotalPages(msg.TotalPages)
		return newModel, nil
	case FormLenSet:
		newModel := m
		newModel.FormLen = msg.FormLen
		return newModel, nil
	case FormSet:
		newModel := m
		newModel.Form = &msg.Form
		return newModel, nil
	case ErrorSet:
		newModel := m
		newModel.Err = msg.Err
		return newModel, nil
	case StatusSet:
		newModel := m
		newModel.Status = msg.Status
		return newModel, nil
	case DialogSet:
		newModel := m
		newModel.Dialog = msg.Dialog
		return newModel, nil
	case DialogActiveSet:
		newModel := m
		newModel.DialogActive = msg.DialogActive
		return newModel, nil
	case RootSet:
		newModel := m
		newModel.Root = msg.Root
		return newModel, nil
	case DatatypeMenuSet:
		newModel := m
		newModel.DatatypeMenu = msg.DatatypeMenu
		return newModel, nil
	case PageMenuSet:
		newModel := m
		newModel.PageMenu = msg.PageMenu
		return newModel, nil
	case DialogReadyOKSet:
		newModel := m
		if newModel.Dialog != nil {
			newModel.Dialog.ReadyOK = msg.Ready
		}
		return newModel, nil

	case DatabaseDeleteEntry:
	case DatabaseCreateEntry:
	case DatabaseUpdateEntry:

	case TablesFetch:
		utility.DefaultLogger.Finfo("Tables Fetch ")
		return m, GetTablesCMD(m.Config)
	case ColumnsFetched:
		return m, tea.Batch(
			LoadingStopCmd(),
			ColumnTypesSetCmd(msg.ColumnTypes),
			ColumnsSetCmd(msg.Columns),
		)

	case TableHeadersRowsFetchedMsg:
		return m, tea.Batch(
			HeadersSetCmd(msg.Headers),
			RowsSetCmd(msg.Rows),
			PaginatorUpdateCmd(m.MaxRows, len(msg.Rows)),
			CursorMaxSetCmd(m.Paginator.ItemsOnPage(len(msg.Rows))),
			LoadingStopCmd(),
		)
	case createFormMsg:
		return m, tea.Batch(
			FormSetCmd(*msg.Form),
			FormLenSetCmd(msg.FieldsCount),
			LoadingStopCmd(),
		)
	case updateFormMsg:
		return m, tea.Batch(
			FormSetCmd(*msg.Form),
			FormLenSetCmd(msg.FieldsCount),
			LoadingStopCmd(),
		)
	case cmsFormMsg:
		return m, tea.Batch(
			FormSetCmd(*msg.Form),
			FormLenSetCmd(msg.FieldsCount),
			LoadingStopCmd(),
		)
	case updateMaxCursorMsg:
		cursorUpdate := func() tea.Msg {
			if m.Cursor > msg.cursorMax-1 {
				return CursorSet{Index: msg.cursorMax - 1}
			}
			return nil
		}()

		cmds := []tea.Cmd{
			CursorMaxSetCmd(msg.cursorMax),
		}

		if cursorUpdate != nil {
			cmds = append(cmds, func() tea.Msg { return cursorUpdate })
		}

		return m, tea.Batch(cmds...)
	case DatatypesFetchedMsg:
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		newMenu := m.BuildDatatypeMenu(msg.data)
		utility.DefaultLogger.Finfo("newMenu", newMenu)

		datatypeMenuLabels := make([]string, 0, len(newMenu))
		for _, item := range newMenu {
			datatypeMenuLabels = append(datatypeMenuLabels, item.Label)
			utility.DefaultLogger.Finfo("item", item)
		}

		return m, tea.Batch(
			DatatypeMenuSetCmd(datatypeMenuLabels),
			PageMenuSetCmd(newMenu),
		)

	case ErrMsg:
		// Handle an error from data fetching.
		return m, tea.Batch(
			ErrorSetCmd(msg.Error),
			LoadingStopCmd(),
		)
	case ShowDialogMsg:
		// Handle showing a dialog
		dialog := NewDialog(msg.Title, msg.Message, msg.ShowCancel)
		return m, tea.Batch(
			DialogSetCmd(&dialog),
			DialogActiveSetCmd(true),
			FocusSetCmd(DIALOGFOCUS),
		)
	case DialogAcceptMsg:
		// Handle dialog accept action
		return m, tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)
	case DialogCancelMsg:
		// Handle dialog cancel action
		return m, tea.Batch(
			DialogActiveSetCmd(false),
			FocusSetCmd(PAGEFOCUS),
		)
	case DialogReadyOK:
		return m, DialogReadyOKSetCmd(true)
	case cms.NewRootMSG:
		return m, RootSetCmd(model.CreateRoot())
	case cms.NewNodeMSG:
		return m, RootSetCmd(model.CreateNode(m.Root, int64(msg.ParentID), int64(msg.DatatypeID), int64(msg.ContentID)))
	case cms.LoadPageMSG:
		// Load page from database using contentID
		return m, func() tea.Msg {
			root, err := model.LoadPageContent(int64(msg.ContentID), *m.Config)
			if err != nil {
				return tea.Batch(
					ErrorSetCmd(err),
					StatusSetCmd(ERROR),
				)()
			}
			return RootSet{Root: root}
		}
	case cms.SavePageMSG:
		// Save page to database
		return m, func() tea.Msg {
			err := model.SavePageContent(m.Root, *m.Config)
			if err != nil {
				return tea.Batch(
					ErrorSetCmd(err),
					StatusSetCmd(ERROR),
				)()
			}
			return nil
		}
	case tea.KeyMsg:
		keyMsg = msg
	default:
		// Check if we need to handle dialog key presses first
		if m.DialogActive && m.Dialog != nil {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				dialog, cmd := m.Dialog.Update(msg)
				m.Dialog = &dialog
				if cmd != nil {
					return m, cmd
				}
			}
		}
	}

	return PageSpecificMsgHandlers(m, nil, keyMsg)

}

func PageSpecificMsgHandlers(m Model, cmd tea.Cmd, msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
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
	case "h", "shift+tab", "backspace":
		if len(m.History) > 0 {
			return m, HistoryPopCmd()
		}
	}
	switch m.Page.Index {
	case HOMEPAGE:
		return m.BasicControls(msg)
	case DATABASEPAGE:
		return m.SelectTable(msg)
	case TABLEPAGE:
		return m.BasicControls(msg)
	case READPAGE:
		return m.DatabaseReadControls(msg)
	case UPDATEPAGE:
		return m.DatabaseReadControls(msg)
	case DEVELOPMENT:
		return DevelopmentInterface(m, msg)
	case DATATYPE:
		return m.DefineDatatypeControls(msg)
	case CONFIGPAGE:
		return m.ConfigControls(msg)

	}
	return m, tea.Batch(cmds...)
}

func (m Model) BasicControls(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	//Navigation
	case "up", "k":
		if m.Cursor > 0 {
			return m, CursorUpCmd()
		}
	case "down", "j":
		if m.Cursor < len(m.PageMenu)-1 {
			return m, CursorDownCmd()
		}
	case "enter", "l":
		// Only proceed if we have menu items
		if len(m.PageMenu) > 0 {
			return m, NavigateToPageCmd(*m.PageMenu[m.Cursor])
		}
	}
	return m, nil

}

func (m Model) SelectTable(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			return m, CursorUpCmd()
		}
	case "down", "j":
		if m.Cursor < len(m.Tables)-1 {
			return m, CursorDownCmd()
		}
	case "enter", "l":
		cmds = append(cmds, SelectTableCmd(m.Tables[m.Cursor]))
	}
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
		cmds = append(cmds, FormAbortedCmd())
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
			d := NewDialog("", "", true)
			m.Dialog = &d
			mg := ShowDialog("Dialog", "test", true)
			cmds = append(cmds, mg)
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	}

	return m, tea.Batch(cmds...)

}

func (m Model) UpdateDatabaseCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.Page = m.Pages[READPAGE]
	}

	if m.Form.State == huh.StateCompleted {
		_ = tea.ClearScreen()
		m.Focus = PAGEFOCUS
		m.Page = m.Pages[READPAGE]
		cmd := FetchTableHeadersRowsCmd(*m.Config, m.Table)
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.Spinner, scmd = m.Spinner.Update(msg)
	cmds = append(cmds, scmd)

	return m, tea.Batch(cmds...)
}

func (m Model) DatabaseReadControls(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < m.CursorMax-1 {
			m.Cursor++
		}
	case "h", "shift+tab", "backspace":
		entry := *m.PopHistory()
		m.Cursor = entry.Cursor
		m.Page = entry.Page
		m.PageMenu = m.Page.Children
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
		m.PushHistory(PageHistory{Page: m.Page, Cursor: m.Cursor})

		recordIndex := (m.PageMod * m.MaxRows) + m.Cursor
		if recordIndex < len(m.Rows) {
			m.Cursor = recordIndex
		}
		cmds = append(cmds, NavigateToPageCmd(*m.PageMenu[m.Cursor]))

	default:
		var pcmd tea.Cmd
		m.Paginator, pcmd = m.Paginator.Update(msg)
		cmds = append(cmds, pcmd)
		cmds = append(cmds, m.UpdateMaxCursor())
	}
	return m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseUpdate(msg tea.KeyMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var rows [][]string
	m.Focus = FORMFOCUS
	switch msg.String() {
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
		cmd := m.CLIUpdate(m.Config, db.DBTable(m.Table))
		cmds = append(cmds, cmd)
	}
	var scmd tea.Cmd
	m.Spinner, scmd = m.Spinner.Update(msg)
	cmds = append(cmds, scmd)

	return m, tea.Batch(cmds...)
}

func (m Model) UpdateDatabaseDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg.String() {
	case "enter", "l":
		err := m.CLIDelete(m.Config, db.StringDBTable(m.Table))
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
	var pcmd tea.Cmd
	m.Paginator, pcmd = m.Paginator.Update(msg)
	cmds = append(cmds, pcmd)
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
