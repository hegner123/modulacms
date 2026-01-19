package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type CmsUpdate struct{}

func NewCmsUpdate() tea.Cmd {
	return func() tea.Msg {
		return CmsUpdate{}
	}
}

func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BuildTreeFromRouteMsg:
		return m, nil
	case CmsDefineDatatypeLoadMsg:
		return m, CmsBuildDefineDatatypeFormCmd()
	case CmsDefineDatatypeReadyMsg:
		return m, nil
	case BuildContentFormMsg:
		// Build dynamic form for content creation
		return m, m.BuildContentFieldsForm(msg.DatatypeID, msg.RouteID)
	case CmsAddNewContentDataMsg:
		// Collect field values from form state
		fieldValues := m.CollectFieldValuesFromForm()

		// Dispatch specialized command using typed methods
		// TODO: Get actual authorID from authenticated user session
		return m, CreateContentWithFieldsCmd(
			m.Config,
			msg.Datatype,
			m.PageRouteId,
			1, // authorID - using default for now
			fieldValues,
		)

	case ContentCreatedMsg:
		// Success path - reload tree and navigate back to content browser
		contentPage := m.PageMap[CONTENT]
		return m, tea.Batch(
			ShowDialog(
				"Success",
				fmt.Sprintf("✓ Created content with %d fields", msg.FieldCount),
				false,
			),
			LogMessageCmd(fmt.Sprintf("ContentData created: ID=%d, RouteID=%d", msg.ContentDataID, msg.RouteID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
			FormCompletedCmd(&contentPage), // Navigate back to content browser
		)

	case ContentCreatedWithErrorsMsg:
		// Partial success path - reload tree even with errors, navigate back
		contentPage := m.PageMap[CONTENT]
		return m, tea.Batch(
			ShowDialog(
				"Warning",
				fmt.Sprintf("⚠ Content created but %d/%d fields failed",
					len(msg.FailedFields),
					msg.CreatedFields+len(msg.FailedFields),
				),
				false,
			),
			LogMessageCmd(fmt.Sprintf("Failed field IDs: %v", msg.FailedFields)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
			FormCompletedCmd(&contentPage), // Navigate back to content browser
		)

	case TreeLoadedMsg:
		// Tree has been reloaded from database
		newModel := m

		// Handle empty tree (route doesn't exist or has no content)
		if msg.RootNode == nil {
			newModel.Root = *NewTreeRoot()
			return newModel, LogMessageCmd(fmt.Sprintf("No content tree found for route %d", msg.RouteID))
		}

		newModel.Root = *msg.RootNode
		return newModel, LogMessageCmd(fmt.Sprintf("Tree reloaded: %d nodes, %d orphans resolved",
			msg.Stats.NodesCount, msg.Stats.OrphansResolved))

	default:
		return m, nil
	}
}
