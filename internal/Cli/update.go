package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	pageInterface       CliInterface = "PageInterface"
	tableInterface      CliInterface = "TableInterface"
	createInterface     CliInterface = "CreateInterface"
	readInterface       CliInterface = "ReadInterface"
	updateInterface     CliInterface = "UpdateInterface"
	deleteInterface     CliInterface = "DeleteInterface"
	updateFormInterface CliInterface = "UpdateFormInterface"
	readSingleInterface CliInterface = "ReadSingleInterface"
	contentInterface    CliInterface = "ContentInterface"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.controller {
	case createInterface:
		return m.UpdateDatabaseCreate(msg)
	case readInterface:
		return m.UpdateDatabaseRead(msg)
	case readSingleInterface:
		return m.UpdateDatabaseRead(msg)
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
	}

	return &m, nil
}
func (m model) UpdateTableSelect(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.TableSelectControls(msg, len(m.tables))
	}
	return &m, nil
}

func (m model) UpdatePageSelect(message tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.PageControls(msg, len(m.menu))
	}
	return &m, nil
}
func (m model) UpdateDatabaseCreate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseCreateControls(message)
}

func (m model) UpdateDatabaseRead(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseReadControls(msg, len(m.rows))
	}
	return &m, nil
}
func (m model) UpdateDatabaseReadSingle(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseReadSingleControls(msg, len(m.rows))
	}
	return &m, nil
}
func (m model) UpdateDatabaseUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseUpdateControls(msg, len(m.rows))
	}
	return &m, nil
}
func (m model) UpdateDatabaseFormUpdate(message tea.Msg) (tea.Model, tea.Cmd) {
	return m.DatabaseUpdateFormControls(message, len(m.rows))
	/*switch msg := message.(type) {
	case tea.KeyMsg:
	}
	return &m, nil*/
}
func (m model) UpdateDatabaseDelete(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseDeleteControls(msg, len(m.rows))
	}
	return &m, nil
}

func (m model) UpdateContent(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ContentControls(msg)

}
