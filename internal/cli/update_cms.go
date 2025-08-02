package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/cli/cms"
	"github.com/hegner123/modulacms/internal/model"
)

type CmsUpdate struct{}

func NewCmsUpdate() tea.Cmd {
	return func() tea.Msg {
		return CmsUpdate{}
	}
}

func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	default:
		return m, nil
	}

}
