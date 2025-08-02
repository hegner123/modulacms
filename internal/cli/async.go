package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

type PageLoadMSG struct{}

func (c Page) PageInit(m Model) tea.Cmd {
	switch c.Index {
	case CMSPAGE:
		r := m.DatabaseRead(m.Config, db.Datatype)
		return r
	default:
		r := PageLoad()
		msg, ok := r.(tea.Cmd)
		if !ok {
			return msg
		}
		return msg

	}

}

func PageLoad() tea.Msg {
	return PageLoadMSG{}
}
