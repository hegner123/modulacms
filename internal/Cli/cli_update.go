package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

var (
	createInterface     CliInterface = "CreateInterface"
	readInterface       CliInterface = "ReadInterface"
	updateInterface     CliInterface = "UpdateInterface"
	deleteInterface     CliInterface = "DeleteInterface"
	tableInterface      CliInterface = "TableInterface"
	pageInterface       CliInterface = "PageInterface"
	inputInterface      CliInterface = "InputInterface"
	updateFormInterface CliInterface = "UpdateFormInterface"
)

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch m.controller {
	case createInterface:
		return m.UpdateDatabaseCreate(message)
	case readInterface:
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
	case inputInterface:
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
func (m model) FormInterface(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.FormInputControl(msg)
	}
	return m, nil

}
func (m model) UpdateDatabaseCreate(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseCreateControls(msg, len(m.tables))
	}
	return m, nil
}

func (m model) UpdateDatabaseRead(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseReadControls(msg, len(*m.rows))
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
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseUpdateFormControls(msg, len(*m.rows))
	}
	return m, nil
}
func (m model) UpdateDatabaseDelete(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		return m.DatabaseDeleteControls(msg, len(*m.rows))
	}
	return m, nil
}
