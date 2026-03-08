package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree"
)

// NewEditDatatypeDialog creates a form dialog for editing a datatype with pre-populated values
func NewEditDatatypeDialog(title string, action FormDialogAction, parents []db.Datatypes, datatype db.Datatypes) FormDialogModel {
	// Create name input with current value
	nameInput := textinput.New()
	nameInput.Placeholder = "Machine name"
	nameInput.CharLimit = 64
	nameInput.Width = 40
	nameInput.SetValue(datatype.Name)
	nameInput.Focus()

	// Create label input with current value
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.SetValue(datatype.Label)

	// Create type input with current value
	typeInput := textinput.New()
	typeInput.Placeholder = "_root"
	typeInput.CharLimit = 32
	typeInput.Width = 40
	typeInput.SetValue(datatype.Type)

	// Build parent options
	parentOptions := []ParentOption{
		{Label: "_root (no parent)", Value: ""},
	}
	selectedParentIndex := 0
	for _, p := range parents {
		// Skip self to prevent circular reference
		if p.DatatypeID == datatype.DatatypeID {
			continue
		}
		parentOptions = append(parentOptions, ParentOption{
			Label: p.Label,
			Value: string(p.DatatypeID),
		})
		// Check if this is the current parent
		if datatype.ParentID.Valid && string(datatype.ParentID.ID) == string(p.DatatypeID) {
			selectedParentIndex = len(parentOptions) - 1
		}
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		EntityID:      string(datatype.DatatypeID),
		NameInput:     nameInput,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parentOptions,
		ParentIndex:   selectedParentIndex,
		focusIndex:    FormDialogFieldName,
	}
}

// NewEditFieldDialog creates a form dialog for editing a field with pre-populated values
func NewEditFieldDialog(title string, action FormDialogAction, field db.Fields) FormDialogModel {
	// Create name input with current value
	nameInput := textinput.New()
	nameInput.Placeholder = "Machine name"
	nameInput.CharLimit = 64
	nameInput.Width = 40
	nameInput.SetValue(field.Name)
	nameInput.Focus()

	// Create label input with current value
	labelInput := textinput.New()
	labelInput.Placeholder = "Display name"
	labelInput.CharLimit = 64
	labelInput.Width = 40
	labelInput.SetValue(field.Label)

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		EntityID:     string(field.FieldID),
		NameInput:    nameInput,
		LabelInput:   labelInput,
		TypeOptions:  TypeOptionsFromRegistry(),
		TypeIndex:    FieldInputTypeIndex(string(field.Type)),
		focusIndex:   FormDialogFieldName,
	}
}

// NewEditRouteDialog creates a form dialog for editing a route with pre-populated values
func NewEditRouteDialog(title string, action FormDialogAction, route db.Routes) FormDialogModel {
	// Create title input with current value (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.SetValue(route.Title)
	titleInput.Focus()

	// Create slug input with current value (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40
	slugInput.SetValue(string(route.Slug))

	return FormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        60,
		Action:       action,
		EntityID:     string(route.RouteID),
		LabelInput:   titleInput,
		TypeInput:    slugInput,
		focusIndex:   FormDialogFieldLabel,
	}
}

// ShowEditDatatypeDialogMsg is the message for showing an edit datatype dialog
type ShowEditDatatypeDialogMsg struct {
	Datatype db.Datatypes
	Parents  []db.Datatypes
}

// ShowEditDatatypeDialogCmd shows an edit dialog for a datatype
func ShowEditDatatypeDialogCmd(datatype db.Datatypes, parents []db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditDatatypeDialogMsg{
			Datatype: datatype,
			Parents:  parents,
		}
	}
}

// ShowEditFieldDialogMsg is the message for showing an edit field dialog
type ShowEditFieldDialogMsg struct {
	Field db.Fields
}

// ShowEditFieldDialogCmd shows an edit dialog for a field
func ShowEditFieldDialogCmd(field db.Fields) tea.Cmd {
	return func() tea.Msg {
		return ShowEditFieldDialogMsg{
			Field: field,
		}
	}
}

// ShowEditRouteDialogMsg is the message for showing an edit route dialog
type ShowEditRouteDialogMsg struct {
	Route db.Routes
}

// ShowEditRouteDialogCmd shows an edit dialog for a route
func ShowEditRouteDialogCmd(route db.Routes) tea.Cmd {
	return func() tea.Msg {
		return ShowEditRouteDialogMsg{
			Route: route,
		}
	}
}

// NewRouteWithContentDialog creates a form dialog for creating a new route with initial content
func NewRouteWithContentDialog(title string, action FormDialogAction, rootDatatypes []db.Datatypes) FormDialogModel {
	// Create title input (uses LabelInput field)
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.Focus()

	// Create slug input (uses TypeInput field)
	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40

	// Build datatype options using ParentOptions carousel
	parentOptions := make([]ParentOption, len(rootDatatypes))
	for i, dt := range rootDatatypes {
		parentOptions[i] = ParentOption{
			Label: dt.Label,
			Value: string(dt.DatatypeID),
		}
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
	}
}

// ShowCreateRouteWithContentDialogMsg is the message for showing a create route with content dialog
type ShowCreateRouteWithContentDialogMsg struct {
	RootDatatypes []db.Datatypes
}

// ShowCreateRouteWithContentDialogCmd shows a dialog to create a new route with initial content
func ShowCreateRouteWithContentDialogCmd(rootDatatypes []db.Datatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowCreateRouteWithContentDialogMsg{
			RootDatatypes: rootDatatypes,
		}
	}
}

// ShowCreateAdminRouteWithContentDialogMsg triggers showing a create admin route with content dialog.
type ShowCreateAdminRouteWithContentDialogMsg struct {
	AdminRootDatatypes []db.AdminDatatypes
}

// ShowCreateAdminRouteWithContentDialogCmd shows a dialog to create a new admin route with initial content.
func ShowCreateAdminRouteWithContentDialogCmd(adminRootDatatypes []db.AdminDatatypes) tea.Cmd {
	return func() tea.Msg {
		return ShowCreateAdminRouteWithContentDialogMsg{
			AdminRootDatatypes: adminRootDatatypes,
		}
	}
}

// NewAdminRouteWithContentDialog creates a form dialog for creating a new admin route with initial content.
func NewAdminRouteWithContentDialog(title string, action FormDialogAction, adminRootDatatypes []db.AdminDatatypes) FormDialogModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Page title"
	titleInput.CharLimit = 128
	titleInput.Width = 40
	titleInput.Focus()

	slugInput := textinput.New()
	slugInput.Placeholder = "url-slug"
	slugInput.CharLimit = 128
	slugInput.Width = 40

	parentOptions := make([]ParentOption, len(adminRootDatatypes))
	for i, dt := range adminRootDatatypes {
		parentOptions[i] = ParentOption{
			Label: dt.Label,
			Value: string(dt.AdminDatatypeID),
		}
	}

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        action,
		LabelInput:    titleInput,
		TypeInput:     slugInput,
		ParentOptions: parentOptions,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldLabel,
	}
}

// ShowInitializeRouteContentDialogMsg is the message for initializing content on an existing route
type ShowInitializeRouteContentDialogMsg struct {
	Route      db.Routes
	DatatypeID string
}

// ShowInitializeRouteContentDialogCmd shows a confirmation dialog to initialize content for a route
func ShowInitializeRouteContentDialogCmd(route db.Routes, datatypeID string) tea.Cmd {
	return func() tea.Msg {
		return ShowInitializeRouteContentDialogMsg{
			Route:      route,
			DatatypeID: datatypeID,
		}
	}
}

// ShowChildDatatypeDialogMsg is the message for showing a child datatype selection dialog
type ShowChildDatatypeDialogMsg struct {
	ParentDatatypeID string
	RouteID          string
	ChildDatatypes   []db.Datatypes
}

// ShowChildDatatypeDialogCmd fetches child datatypes and shows a selection dialog
func ShowChildDatatypeDialogCmd(parentDatatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return FetchChildDatatypesMsg{
			ParentDatatypeID: parentDatatypeID,
			RouteID:          routeID,
		}
	}
}

// FetchChildDatatypesMsg triggers fetching child datatypes for a parent
type FetchChildDatatypesMsg struct {
	ParentDatatypeID types.DatatypeID
	RouteID          types.RouteID
}

// ChildDatatypeSelectedMsg is sent when a child datatype is selected from the dialog
type ChildDatatypeSelectedMsg struct {
	DatatypeID types.DatatypeID
	RouteID    types.RouteID
}

// ChildDatatypeSelectedCmd creates a command that returns a ChildDatatypeSelectedMsg
func ChildDatatypeSelectedCmd(datatypeID types.DatatypeID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return ChildDatatypeSelectedMsg{
			DatatypeID: datatypeID,
			RouteID:    routeID,
		}
	}
}

// NewChildDatatypeDialog creates a dialog for selecting a child datatype
func NewChildDatatypeDialog(title string, childDatatypes []db.Datatypes, routeID string) FormDialogModel {
	// Build parent options from child datatypes
	parents := make([]ParentOption, 0, len(childDatatypes))
	for _, dt := range childDatatypes {
		parents = append(parents, ParentOption{
			Label: dt.Label,
			Value: string(dt.DatatypeID),
		})
	}

	// Initialize text inputs even though they're not displayed
	// This prevents nil pointer panics when updateFocus is called
	labelInput := textinput.New()
	labelInput.Placeholder = ""
	typeInput := textinput.New()
	typeInput.Placeholder = ""

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         50,
		Action:        FORMDIALOGCHILDDATATYPE,
		EntityID:      routeID,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: parents,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldParent, // Start on selection
	}
}

// =============================================================================
// MOVE CONTENT DIALOG
// =============================================================================

// ShowMoveContentDialogMsg triggers showing the move content dialog
type ShowMoveContentDialogMsg struct {
	SourceNode   *tree.Node
	RouteID      types.RouteID
	ValidTargets []ParentOption
}

// MoveContentRequestMsg triggers the actual content move operation
type MoveContentRequestMsg struct {
	SourceContentID types.ContentID
	TargetContentID types.ContentID
	RouteID         types.RouteID
}

// ContentMovedMsg is sent after content is successfully moved
type ContentMovedMsg struct {
	SourceContentID types.ContentID
	TargetContentID types.ContentID
	RouteID         types.RouteID
	AdminMode       bool
}

// ShowMoveContentDialogCmd creates a command to show the move content dialog
func ShowMoveContentDialogCmd(node *tree.Node, routeID types.RouteID, targets []ParentOption) tea.Cmd {
	return func() tea.Msg {
		return ShowMoveContentDialogMsg{
			SourceNode:   node,
			RouteID:      routeID,
			ValidTargets: targets,
		}
	}
}

// MoveContentCmd creates a command that returns a MoveContentRequestMsg
func MoveContentCmd(sourceID, targetID types.ContentID, routeID types.RouteID) tea.Cmd {
	return func() tea.Msg {
		return MoveContentRequestMsg{
			SourceContentID: sourceID,
			TargetContentID: targetID,
			RouteID:         routeID,
		}
	}
}

// NewMoveContentDialog creates a dialog for selecting a move target
func NewMoveContentDialog(title string, sourceContentID string, routeID string, targets []ParentOption) FormDialogModel {
	// Initialize text inputs even though they're not displayed
	// This prevents nil pointer panics when updateFocus is called
	labelInput := textinput.New()
	labelInput.Placeholder = ""
	typeInput := textinput.New()
	typeInput.Placeholder = ""

	return FormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         50,
		Action:        FORMDIALOGMOVECONTENT,
		EntityID:      sourceContentID + "|" + routeID,
		LabelInput:    labelInput,
		TypeInput:     typeInput,
		ParentOptions: targets,
		ParentIndex:   0,
		focusIndex:    FormDialogFieldParent, // Start on selection
	}
}

// =============================================================================
// SINGLE CONTENT FIELD EDIT DIALOG
// =============================================================================

// ShowEditSingleFieldDialogMsg triggers showing a single-field edit dialog.
type ShowEditSingleFieldDialogMsg struct {
	Field      ContentFieldDisplay
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowEditSingleFieldDialogCmd creates a command to show a single-field edit dialog.
func ShowEditSingleFieldDialogCmd(cf ContentFieldDisplay, contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowEditSingleFieldDialogMsg{
			Field:      cf,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// EditSingleFieldAcceptMsg carries acceptance data from a single-field edit dialog.
type EditSingleFieldAcceptMsg struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	FieldID        types.FieldID
	NewValue       string
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
}

// =============================================================================
// ADD CONTENT FIELD DIALOG (picker for multiple missing fields)
// =============================================================================

// ShowAddContentFieldDialogMsg triggers showing an add-field picker dialog.
type ShowAddContentFieldDialogMsg struct {
	Options    []huh.Option[string]
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowAddContentFieldDialogCmd creates a command to show an add-field picker.
func ShowAddContentFieldDialogCmd(options []huh.Option[string], contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowAddContentFieldDialogMsg{
			Options:    options,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// DELETE CONTENT FIELD DIALOG
// =============================================================================

// DeleteContentFieldContext stores context for a content field deletion operation.
// AdminMode selects admin vs regular delete flow.
type DeleteContentFieldContext struct {
	ContentFieldID types.ContentFieldID
	ContentID      types.ContentID
	RouteID        types.RouteID
	DatatypeID     types.NullableDatatypeID
	AdminMode      bool
}

// ShowDeleteContentFieldDialogMsg triggers showing a delete content field confirmation dialog.
type ShowDeleteContentFieldDialogMsg struct {
	Field      ContentFieldDisplay
	ContentID  types.ContentID
	RouteID    types.RouteID
	DatatypeID types.NullableDatatypeID
}

// ShowDeleteContentFieldDialogCmd creates a command to show a delete content field dialog.
func ShowDeleteContentFieldDialogCmd(cf ContentFieldDisplay, contentID types.ContentID, routeID types.RouteID, datatypeID types.NullableDatatypeID) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteContentFieldDialogMsg{
			Field:      cf,
			ContentID:  contentID,
			RouteID:    routeID,
			DatatypeID: datatypeID,
		}
	}
}

// =============================================================================
// EXTERNAL EDITOR SUPPORT
// =============================================================================

// EditorFinishedMsg is sent after $EDITOR exits with the edited content.
type EditorFinishedMsg struct {
	FieldIndex int
	Content    string
	Err        error
}

// editorFileExtension returns a file extension hint for the given widget type.
func editorFileExtension(widget string) string {
	switch widget {
	case "markdown":
		return ".md"
	case "json-editor":
		return ".json"
	case "code-editor":
		return ".txt"
	case "rich-text":
		return ".html"
	default:
		return ".txt"
	}
}

// sshExecCommand wraps an exec.Cmd for use with tea.Exec, routing stderr to
// the same writer as stdout. This is necessary for SSH sessions where
// Bubbletea's default exec hardcodes stderr to os.Stderr (the server's
// stderr), which disconnects interactive editors like nvim from the SSH
// client's terminal.
type sshExecCommand struct {
	*exec.Cmd
}

func (c *sshExecCommand) SetStdin(r io.Reader) {
	if c.Stdin == nil {
		c.Stdin = r
	}
}

func (c *sshExecCommand) SetStdout(w io.Writer) {
	if c.Stdout == nil {
		c.Stdout = w
	}
	// Also route stderr to the SSH session writer so the editor's
	// full TUI output reaches the client.
	if c.Stderr == nil {
		c.Stderr = w
	}
}

func (c *sshExecCommand) SetStderr(w io.Writer) {
	// No-op: stderr was already wired to the SSH session in SetStdout.
	// This prevents Bubbletea's default of os.Stderr from overriding it.
}

// prepareEditorCmd synchronously creates a temp file with the current value,
// then returns a tea.Exec cmd that launches $EDITOR with stdin, stdout, and
// stderr all routed through the SSH session's PTY. When the editor exits,
// the callback reads the file and returns an EditorFinishedMsg.
// Returns nil if temp file creation fails. Logger may be nil.
func prepareEditorCmd(fieldIndex int, currentValue string, widget string, logger Logger) tea.Cmd {
	ext := editorFileExtension(widget)
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] file extension for widget %q: %s", widget, ext))
	}

	tmpFile, err := os.CreateTemp("", "modula-*"+ext)
	if err != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to create temp file with extension %s", ext), err)
		}
		return nil
	}
	tmpPath := tmpFile.Name()
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] created temp file: %s", tmpPath))
	}

	if _, writeErr := tmpFile.WriteString(currentValue); writeErr != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to write current value (%d bytes) to temp file %s", len(currentValue), tmpPath), writeErr)
		}
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil
	}
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] wrote %d bytes to temp file %s", len(currentValue), tmpPath))
	}

	if closeErr := tmpFile.Close(); closeErr != nil {
		if logger != nil {
			logger.Ferror(fmt.Sprintf("[editor] failed to close temp file %s", tmpPath), closeErr)
		}
		os.Remove(tmpPath)
		return nil
	}

	editor := editorCommand()
	editorParts := strings.Fields(editor)
	if len(editorParts) == 0 {
		if logger != nil {
			logger.Ferror("[editor] resolved editor command is empty after splitting", fmt.Errorf("empty editor command: %q", editor))
		}
		os.Remove(tmpPath)
		return nil
	}
	editorArgs := append(editorParts[1:], tmpPath)
	if logger != nil {
		logger.Finfo(fmt.Sprintf("[editor] resolved editor command: %q (binary: %q, args: %v)", editor, editorParts[0], editorArgs))
		logger.Finfo(fmt.Sprintf("[editor] launching (via sshExecCommand): %s %s", editorParts[0], strings.Join(editorArgs, " ")))
	}
	c := exec.Command(editorParts[0], editorArgs...)

	return tea.Exec(&sshExecCommand{Cmd: c}, func(procErr error) tea.Msg {
		defer os.Remove(tmpPath)
		if procErr != nil {
			if logger != nil {
				logger.Ferror(fmt.Sprintf("[editor] editor process exited with error for field %d, temp file %s", fieldIndex, tmpPath), procErr)
			}
			return EditorFinishedMsg{FieldIndex: fieldIndex, Err: procErr}
		}
		if logger != nil {
			logger.Finfo(fmt.Sprintf("[editor] editor process exited successfully for field %d, reading back temp file %s", fieldIndex, tmpPath))
		}
		data, readErr := os.ReadFile(tmpPath)
		if readErr != nil {
			if logger != nil {
				logger.Ferror(fmt.Sprintf("[editor] failed to read temp file %s after editor exit", tmpPath), readErr)
			}
			return EditorFinishedMsg{FieldIndex: fieldIndex, Err: readErr}
		}
		if logger != nil {
			logger.Finfo(fmt.Sprintf("[editor] read %d bytes from temp file %s, returning EditorFinishedMsg for field %d", len(data), tmpPath, fieldIndex))
		}
		return EditorFinishedMsg{FieldIndex: fieldIndex, Content: string(data)}
	})
}
