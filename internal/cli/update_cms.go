package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
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
	case CreateRouteFromDialogRequestMsg:
		return m, m.HandleCreateRouteFromDialog(msg)
	case UpdateDatatypeFromDialogRequestMsg:
		return m, m.HandleUpdateDatatypeFromDialog(msg)
	case UpdateFieldFromDialogRequestMsg:
		return m, m.HandleUpdateFieldFromDialog(msg)
	case UpdateRouteFromDialogRequestMsg:
		return m, m.HandleUpdateRouteFromDialog(msg)
	case CreateRouteWithContentRequestMsg:
		return m, m.HandleCreateRouteWithContent(msg)
	case InitializeRouteContentRequestMsg:
		return m, m.HandleInitializeRouteContent(msg)
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
	case ChildDatatypeSelectedMsg:
		// User selected a child datatype from the dialog - fetch fields and show content form
		m.Logger.Finfo(fmt.Sprintf("ChildDatatypeSelectedMsg received: DatatypeID=%s, RouteID=%s", msg.DatatypeID, msg.RouteID))
		selectedNode := m.Root.NodeAtIndex(m.Cursor)
		var parentContentID types.NullableContentID
		if selectedNode != nil && selectedNode.Instance != nil {
			parentContentID = types.NullableContentID{ID: selectedNode.Instance.ContentDataID, Valid: true}
			m.Logger.Finfo(fmt.Sprintf("Parent content ID: %s", parentContentID.ID))
		}
		return m, FetchContentFieldsCmd(msg.DatatypeID, msg.RouteID, parentContentID, "New Content")
	case CreateContentFromDialogRequestMsg:
		// Create content from dialog values using authenticated user
		return m, m.HandleCreateContentFromDialog(msg, m.UserID)
	case FetchContentForEditMsg:
		// Fetch content fields for editing - this runs in background and shows edit dialog
		return m, m.HandleFetchContentForEdit(msg)
	case UpdateContentFromDialogRequestMsg:
		// Update content from dialog values using authenticated user
		return m, m.HandleUpdateContentFromDialog(msg, m.UserID)
	case MoveContentRequestMsg:
		return m, m.HandleMoveContent(msg)
	case ContentMovedMsg:
		return m, tea.Batch(
			LoadingStopCmd(),
			ShowDialog("Success", "Content moved successfully", false),
			LogMessageCmd(fmt.Sprintf("Content moved: %s -> %s", msg.SourceContentID, msg.TargetContentID)),
			ReloadContentTreeCmd(m.Config, msg.RouteID),
		)
	case DeleteContentRequestMsg:
		// Delete content
		return m, m.HandleDeleteContent(msg)
	case DeleteDatatypeRequestMsg:
		// Delete datatype and its junction records
		return m, m.HandleDeleteDatatype(msg)
	case DeleteFieldRequestMsg:
		// Delete field and its junction record
		return m, m.HandleDeleteField(msg)
	case DeleteRouteRequestMsg:
		// Delete route
		return m, m.HandleDeleteRoute(msg)
	case MediaUploadStartMsg:
		return m, m.HandleMediaUpload(msg)
	case MediaUploadedMsg:
		return m, tea.Batch(
			MediaFetchCmd(),
			ShowDialog("Upload Complete", fmt.Sprintf("'%s' uploaded successfully.", msg.Name), false),
		)
	case DeleteMediaRequestMsg:
		// Delete media item
		return m, m.HandleDeleteMedia(msg)
	case CreateUserFromDialogRequestMsg:
		// Create user from form dialog
		return m, m.HandleCreateUserFromDialog(msg)
	case UpdateUserFromDialogRequestMsg:
		// Update user from form dialog
		return m, m.HandleUpdateUserFromDialog(msg)
	case DeleteUserRequestMsg:
		// Delete user
		return m, m.HandleDeleteUser(msg)
	case CmsAddNewContentDataMsg:
		// Collect field values from form state
		fieldValues := m.CollectFieldValuesFromForm()

		// Dispatch specialized command using typed methods with authenticated user
		return m, CreateContentWithFieldsCmd(
			m.Config,
			msg.Datatype,
			m.PageRouteId,
			m.UserID,
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

	case ContentDeletedMsg:
		// Content deleted successfully - reload tree and show success
		newModel := m
		// Reset cursor if it's beyond the new tree size
		newModel.Cursor = 0
		return newModel, tea.Batch(
			LoadingStopCmd(),
			ShowDialog(
				"Success",
				"Content deleted successfully",
				false,
			),
			LogMessageCmd(fmt.Sprintf("Content deleted: ID=%s", msg.ContentID)),
			ReloadContentTreeCmd(m.Config, types.RouteID(msg.RouteID)),
		)

	case TreeLoadedMsg:
		// Tree has been reloaded from database
		newModel := m

		// Handle empty tree (route doesn't exist or has no content)
		if msg.RootNode == nil {
			newModel.Root = *tree.NewRoot()
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
