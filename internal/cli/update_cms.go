package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
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
		// Show form dialog instead of navigating to separate page
		return m, ShowFormDialogCmd(FORMDIALOGCREATEDATATYPE, "New Datatype", m.AllDatatypes)
	case CreateDatatypeFromDialogRequestMsg:
		return m, m.HandleCreateDatatypeFromDialog(msg)
	case CreateFieldFromDialogRequestMsg:
		return m, m.HandleCreateFieldFromDialog(msg)
	case CmsEditDatatypeLoadMsg:
		return m, CmsEditDatatypeFormCmd(msg.Datatype)
	case DatatypeUpdateSaveMsg:
		d := m.DB
		datatypeID := msg.DatatypeID
		parentStr := msg.Parent
		label := msg.Label
		dtType := msg.Type
		return m, func() tea.Msg {
			parentID := types.NullableContentID{}
			if parentStr != "" {
				parentID = types.NullableContentID{
					ID:    types.ContentID(parentStr),
					Valid: true,
				}
			}
			params := db.UpdateDatatypeParams{
				DatatypeID:   datatypeID,
				ParentID:     parentID,
				Label:        label,
				Type:         dtType,
				DateModified: types.TimestampNow(),
			}
			_, err := d.UpdateDatatype(params)
			if err != nil {
				return DatatypeUpdateFailedMsg{Error: err}
			}
			return DatatypeUpdatedMsg{DatatypeID: datatypeID, Label: label}
		}
	case DatatypeUpdatedMsg:
		datatypesPage := m.PageMap[DATATYPES]
		return m, tea.Batch(
			LoadingStartCmd(),
			LogMessageCmd(fmt.Sprintf("Datatype updated: %s", msg.Label)),
			AllDatatypesFetchCmd(),
			FormCompletedCmd(&datatypesPage),
		)
	case DatatypeUpdateFailedMsg:
		return m, LogMessageCmd(fmt.Sprintf("Datatype update failed: %v", msg.Error))
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
		// Using a default UserID - in production this should come from the session
		defaultAuthorID := types.UserID("01JTRBZ0000000000000000001") // Placeholder author ID
		return m, CreateContentWithFieldsCmd(
			m.Config,
			msg.Datatype,
			m.PageRouteId,
			defaultAuthorID,
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
			LogMessageCmd(fmt.Sprintf("ContentData created: ID=%s, RouteID=%s", msg.ContentDataID, msg.RouteID)),
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
			return newModel, tea.Batch(
				LoadingStopCmd(),
				LogMessageCmd(fmt.Sprintf("No content tree found for route %s", msg.RouteID)),
			)
		}

		newModel.Root = *msg.RootNode
		return newModel, tea.Batch(
			LoadingStopCmd(),
			LogMessageCmd(fmt.Sprintf("Tree reloaded: %d nodes, %d orphans resolved",
				msg.Stats.NodesCount, msg.Stats.OrphansResolved)),
		)

	default:
		return m, nil
	}
}
