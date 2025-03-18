package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

var (
    pageInterface       CliInterface = "PageInterface"
    tableInterface      CliInterface = "TableInterface"
	createInterface     CliInterface = "CreateInterface"
	readInterface       CliInterface = "ReadInterface"
	updateInterface     CliInterface = "UpdateInterface"
	deleteInterface     CliInterface = "DeleteInterface"
	updateFormInterface CliInterface = "UpdateFormInterface"
	readSingleInterface CliInterface = "ReadSingleInterface"
)

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	// First handle any form-specific messages
	switch message.(type) {
	case formCompletedMsg:
		// Process form data here (save to database, etc.)
		m.focus = PAGEFOCUS
		m.page = m.pages[Update]
		m.controller = m.page.Controller
		return m, nil
	case formCancelledMsg:
		m.focus = PAGEFOCUS
		m.page = m.pages[Update]
		m.controller = m.page.Controller
		return m, nil
	}
	switch m.controller {
	case createInterface:
		return m.UpdateDatabaseCreate(message)
	case readInterface:
		return m.UpdateDatabaseRead(message)
	case readSingleInterface:
		return m.UpdateDatabaseRead(message)
	case updateInterface:
		return m.UpdateDatabaseUpdate(message)
	case updateFormInterface:
		return m.UpdateDatabaseFormUpdate(message)
	case deleteInterface:
		return m.UpdateDatabaseDelete(message)
	case pageInterface:
		return m.UpdatePageSelect(message)
	case tableInterface:
		return m.UpdateTableSelect(message)
	}
	return m, nil
}
func (m model) UpdateTableSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.TableSelectControls(msg, len(m.tables))
	}
	return m, nil
}

func (m model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.PageControls(msg, len(m.menu))
	}
	return m, nil
}
func (m model) UpdateDatabaseCreate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseCreateControls(message)
}

func (m model) UpdateDatabaseRead(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseReadControls(msg, len(*m.rows))
	}
	return m, nil
}
func (m model) UpdateDatabaseReadSingle(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseReadSingleControls(msg, len(*m.rows))
	}
	return m, nil
}
func (m model) UpdateDatabaseUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseUpdateControls(msg, len(*m.rows))
	}
	return m, nil
}
func (m model) UpdateDatabaseFormUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseUpdateFormControls(message, len(*m.rows))
	/*switch msg := message.(type) {
	case tea.KeyMsg:
	}
	return m, nil*/
}
func (m model) UpdateDatabaseDelete(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseDeleteControls(msg, len(*m.rows))
	}
	return m, nil
}
