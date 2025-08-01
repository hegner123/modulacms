package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/cli/cms"
	"github.com/hegner123/modulacms/internal/model"
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
	case tableFetchedMsg:
		// Data fetched successfully; update the model.
		utility.DefaultLogger.Finfo("tableFetchedMsg returned")
		m.Tables = msg.Tables
		m.Loading = false
		return m, nil
	case columnFetchedMsg:
		m.Columns = msg.Columns
		m.ColumnTypes = msg.ColumnTypes
		m.Loading = false
		return m, nil
	case headersRowsFetchedMsg:
		m.Headers = msg.Headers
		m.Rows = msg.Rows
		m.Paginator.PerPage = m.MaxRows
		m.Paginator.SetTotalPages(len(m.Rows))
		m.CursorMax = m.Paginator.ItemsOnPage(len(m.Rows))
		m.Loading = false
		return m, nil
	case createFormMsg:
		m.Form = msg.Form
		m.FormLen = msg.FieldsCount
		m.Loading = false
		return m, nil
	case updateFormMsg:
		m.Form = msg.Form
		m.FormLen = msg.FieldsCount
		m.Loading = false
		return m, nil
	case cmsFormMsg:
		m.Form = msg.Form
		m.FormLen = msg.FieldsCount
		m.Loading = false
		return m, nil
	case updateMaxCursorMsg:
		m.CursorMax = msg.cursorMax
		if m.Cursor > m.CursorMax-1 {
			m.Cursor = m.CursorMax - 1
		}
	case ErrMsg:
		// Handle an error from data fetching.
		m.Err = msg.Error
		m.Loading = false
		return m, nil
	case ShowDialogMsg:
		// Handle showing a dialog
		dialog := NewDialog(msg.Title, msg.Message, msg.ShowCancel)
		m.Dialog = &dialog
		m.DialogActive = true
		m.Focus = DIALOGFOCUS
		return m, nil
	case DialogAcceptMsg:
		// Handle dialog accept action
		m.DialogActive = false
		m.Focus = PAGEFOCUS
		return m, nil
	case DialogCancelMsg:
		// Handle dialog cancel action
		m.DialogActive = false
		m.Focus = PAGEFOCUS
		return m, nil
	case DialogReadyOK:
		m.Dialog.ReadyOK = true
		return m, nil
	case cms.NewRootMSG:
		m.Root = model.CreateRoot()
		return m, nil
	case cms.NewNodeMSG:
		m.Root = model.CreateNode(m.Root, int64(msg.ParentID), int64(msg.DatatypeID), int64(msg.ContentID))
		return m, nil
	case cms.LoadPageMSG:
		// Load page from database using contentID
		root, err := model.LoadPageContent(int64(msg.ContentID), *m.Config)
		if err != nil {
			m.Err = err
			m.Status = ERROR
		} else {
			m.Root = root
		}
		return m, nil
	case cms.SavePageMSG:
		// Save page to database
		err := model.SavePageContent(m.Root, *m.Config)
		if err != nil {
			m.Err = err
			m.Status = ERROR
		}
		return m, nil
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

		switch m.Controller {
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
			m.Dialog = &d
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
		m, tcmd = m.TableSelectControls(msg, len(m.Tables))
		cmds = append(cmds, tcmd)
	default:
		// Only update spinner if we're in a loading state
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return &m, tea.Batch(cmds...)
}

func (m Model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.PageControls(msg, len(m.PageMenu))
		cmds = append(cmds, tcmd)
	default:
		// Only update spinner if we're in a loading state
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
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
		m, tcmd = m.DatabaseReadControls(msg, len(m.Rows))
		cmds = append(cmds, tcmd)
	default:
		// Only update spinner if we're in a loading state
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseReadSingle(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.DatabaseReadSingleControls(msg, len(m.Rows))
		cmds = append(cmds, tcmd)
	default:
		// Don't update spinner for other message types to prevent constant re-rendering
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := message.(type) {
	case tea.KeyMsg:
		var tcmd tea.Cmd
		m, tcmd = m.DatabaseUpdateControls(msg, len(m.Rows))
		cmds = append(cmds, tcmd)
	default:
		// Only update spinner if we're in a loading state
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return &m, tea.Batch(cmds...)
}
func (m Model) UpdateDatabaseFormUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseUpdateFormControls(message, len(m.Rows))
}
func (m Model) UpdateDatabaseDelete(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseDeleteControls(msg, len(m.Rows))
	}
	return &m, nil
}

func (m Model) UpdateContent(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ContentControls(msg)

}

func (m Model) UpdateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ConfigControls(msg)

}