package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/utility"
)

const (
	developmentInterface CliInterface = "DevelopmentInterface"
	pageInterface        CliInterface = "PageInterface"
	tableInterface       CliInterface = "TableInterface"
	createInterface      CliInterface = "CreateInterface"
	readInterface        CliInterface = "ReadInterface"
	updateInterface      CliInterface = "UpdateInterface"
	deleteInterface      CliInterface = "DeleteInterface"
	updateFormInterface  CliInterface = "UpdateFormInterface"
	readSingleInterface  CliInterface = "ReadSingleInterface"
	contentInterface     CliInterface = "ContentInterface"
	configInterface      CliInterface = "ConfigInterface"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		headerHeight := lipgloss.Height(m.headerView() + RenderTitle(m.titles[m.titleFont]) + RenderHeading(m.header))
		footerHeight := lipgloss.Height(m.footerView() + RenderFooter(m.footer))
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.YPosition = headerHeight
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMarginHeight - 10
		}
	case tableFetchedMsg:
		// Data fetched successfully; update the model.
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		m.tables = msg.Tables
		m.loading = false
		return m, nil
	case columnFetchedMsg:
		m.columns = msg.Columns
		m.columnTypes = msg.ColumnTypes
		m.loading = false
		return m, nil
	case headersRowsFetchedMsg:
		m.headers = msg.Headers
		m.rows = msg.Rows
		m.paginator.PerPage = m.maxRows
		m.paginator.SetTotalPages(len(m.rows))
		m.cursorMax = m.paginator.ItemsOnPage(len(m.rows))
		m.loading = false
		return m, nil
	case createFormMsg:
		m.form = msg.Form
		m.formLen = msg.FieldsCount
		m.loading = false
		return m, nil
	case updateFormMsg:
		m.form = msg.Form
		m.formLen = msg.FieldsCount
		m.loading = false
		return m, nil
	case cmsFormMsg:
		m.form = msg.Form
		m.formLen = msg.FieldsCount
		m.loading = false
		return m, nil
	case updateMaxCursorMsg:
		m.cursorMax = msg.cursorMax
		if m.cursor > m.cursorMax-1 {
			m.cursor = m.cursorMax - 1
		}
	case ErrMsg:
		// Handle an error from data fetching.
		m.err = msg.Error
		m.loading = false
		return m, nil
	case ShowDialogMsg:
		// Handle showing a dialog
		dialog := NewDialog(msg.Title, msg.Message, msg.ShowCancel)
		m.dialog = &dialog
		m.dialogActive = true
		m.focus = DIALOGFOCUS
		return m, nil
	case DialogAcceptMsg:
		// Handle dialog accept action
		m.dialogActive = false
		m.focus = PAGEFOCUS
		return m, nil
	case DialogCancelMsg:
		// Handle dialog cancel action
		m.dialogActive = false
		m.focus = PAGEFOCUS
		return m, nil
	default:
		// Check if we need to handle dialog key presses first
		if m.dialogActive && m.dialog != nil {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				dialog, cmd := m.dialog.Update(msg)
				m.dialog = &dialog
				if cmd != nil {
					return m, cmd
				}
			}
		}

		switch m.controller {
		case developmentInterface:
			return m.DevelopmentInterface(msg)
		case createInterface:
			return m.UpdateDatabaseCreate(msg)
		case readInterface:
			return m.UpdateDatabaseRead(msg)
		case readSingleInterface:
			return m.UpdateDatabaseReadSingle(msg)
		case updateInterface:
			return m.UpdateDatabaseUpdate(msg)
		case updateFormInterface:
			return m.UpdateDatabaseFormUpdate(msg)
		case deleteInterface:
			return m.UpdateDatabaseDelete(msg)
		case pageInterface:
			return m.UpdatePageSelect(msg)
		case tableInterface:
			return m.UpdateTableSelect(msg)
		case contentInterface:
			return m.UpdateContent(msg)
		case configInterface:
			return m.UpdateConfig(msg)

		}
	}
	return m, nil

}

func (m *Model) DevelopmentInterface(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		//Exit
		case "d":
			d := NewDialog("", "", true)
			m.dialog = &d
			mg := ShowDialog("Dialog", "test", true)
			cmds = append(cmds, mg)
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	}

	return m, tea.Batch(cmds...)

}
func (m Model) UpdateTableSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.TableSelectControls(msg, len(m.tables))
		cmds = append(cmds, tcmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return &m, tea.Batch(cmds...)
}

func (m Model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.PageControls(msg, len(m.pageMenu))
		cmds = append(cmds, tcmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseCreate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseCreateControls(message)
}

func (m Model) UpdateDatabaseRead(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.DatabaseReadControls(msg, len(m.rows))
		cmds = append(cmds, tcmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseReadSingle(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.DatabaseReadSingleControls(msg, len(m.rows))
		cmds = append(cmds, tcmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.DatabaseUpdateControls(msg, len(m.rows))
		cmds = append(cmds, tcmd)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseFormUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseUpdateFormControls(message, len(m.rows))
}
func (m Model) UpdateDatabaseDelete(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseDeleteControls(msg, len(m.rows))
	}
	return &m, nil
}

func (m Model) UpdateContent(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ContentControls(msg)

}

func (m Model) UpdateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ConfigControls(msg)

}
