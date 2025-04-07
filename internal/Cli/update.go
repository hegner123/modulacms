package cli

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	configInterface     CliInterface = "ConfigInterface"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView() + RenderTitle(m.titles[m.titleFont]) + RenderHeading(m.header))
		footerHeight := lipgloss.Height(m.footerView() + RenderFooter(m.footer))
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.YPosition = headerHeight
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMarginHeight - 10
		}
	}

	switch m.controller {
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
		return m.PageControls(msg, len(m.pageMenu))
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

func (m model) UpdateConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.ConfigControls(msg)

}
