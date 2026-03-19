package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// ADMIN MEDIA FOLDER NAME DIALOG (CREATE / RENAME)
// =============================================================================

// AdminMediaFolderNameDialogModel is a simple single-input dialog for creating or renaming admin folders.
type AdminMediaFolderNameDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	IsRename bool // true for rename, false for create

	NameInput textinput.Model

	// Context for the operation
	FolderID types.AdminMediaFolderID         // set for rename only
	ParentID types.NullableAdminMediaFolderID // set for create only

	focusIndex int // 0=input, 1=cancel, 2=confirm
}

const (
	amfNameFocusInput   = 0
	amfNameFocusCancel  = 1
	amfNameFocusConfirm = 2
	amfNameMaxFocus     = 2
)

// NewCreateAdminMediaFolderDialog creates a dialog for creating a new admin media folder.
func NewCreateAdminMediaFolderDialog(parentID types.NullableAdminMediaFolderID) AdminMediaFolderNameDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "folder name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.Focus()

	return AdminMediaFolderNameDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        "New Admin Folder",
		Width:        50,
		IsRename:     false,
		NameInput:    nameInput,
		ParentID:     parentID,
		focusIndex:   amfNameFocusInput,
	}
}

// NewRenameAdminMediaFolderDialog creates a dialog for renaming an existing admin media folder.
func NewRenameAdminMediaFolderDialog(folderID types.AdminMediaFolderID, currentName string) AdminMediaFolderNameDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "new name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.SetValue(currentName)
	nameInput.Focus()

	return AdminMediaFolderNameDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        "Rename Admin Folder",
		Width:        50,
		IsRename:     true,
		NameInput:    nameInput,
		FolderID:     folderID,
		focusIndex:   amfNameFocusInput,
	}
}

// Update handles user input for the admin folder name dialog.
func (d *AdminMediaFolderNameDialogModel) Update(msg tea.Msg) (AdminMediaFolderNameDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.focusIndex++
			if d.focusIndex > amfNameMaxFocus {
				d.focusIndex = amfNameFocusInput
			}
			d.updateFocus()
			return *d, nil
		case "shift+tab", "up":
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = amfNameMaxFocus
			}
			d.updateFocus()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return AdminMediaFolderNameDialogCancelMsg{} }
		case "enter":
			if d.focusIndex == amfNameFocusCancel {
				return *d, func() tea.Msg { return AdminMediaFolderNameDialogCancelMsg{} }
			}
			if d.focusIndex == amfNameFocusConfirm || d.focusIndex == amfNameFocusInput {
				name := strings.TrimSpace(d.NameInput.Value())
				if name == "" {
					return *d, nil // no-op on empty name
				}
				if d.IsRename {
					folderID := d.FolderID
					return *d, func() tea.Msg {
						return RenameAdminMediaFolderRequestMsg{
							FolderID: folderID,
							NewName:  name,
						}
					}
				}
				parentID := d.ParentID
				return *d, func() tea.Msg {
					return CreateAdminMediaFolderRequestMsg{
						Name:     name,
						ParentID: parentID,
					}
				}
			}
			return *d, nil
		default:
			if d.focusIndex == amfNameFocusInput {
				var cmd tea.Cmd
				d.NameInput, cmd = d.NameInput.Update(msg)
				return *d, cmd
			}
		}
	}
	return *d, nil
}

func (d *AdminMediaFolderNameDialogModel) updateFocus() {
	if d.focusIndex == amfNameFocusInput {
		d.NameInput.Focus()
	} else {
		d.NameInput.Blur()
	}
}

// OverlayUpdate implements ModalOverlay for AdminMediaFolderNameDialogModel.
func (d *AdminMediaFolderNameDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayTick implements OverlayTicker for cursor blinking.
func (d *AdminMediaFolderNameDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	if d.focusIndex == amfNameFocusInput {
		var cmd tea.Cmd
		d.NameInput, cmd = d.NameInput.Update(msg)
		return d, cmd
	}
	return d, nil
}

// OverlayView implements ModalOverlay for AdminMediaFolderNameDialogModel.
func (d *AdminMediaFolderNameDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the admin folder name dialog.
func (d *AdminMediaFolderNameDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	// Header
	header := d.titleStyle.Render(d.Title)

	// Name input field
	nameLabel := d.labelStyle.Render("Name")
	nameInputStyle := d.inputStyle
	if d.focusIndex == amfNameFocusInput {
		nameInputStyle = d.focusedInputStyle
	}
	nameField := nameInputStyle.Width(innerW).Render(d.NameInput.View())

	// Buttons
	cancelBtn := d.renderAMFButton("Cancel", d.focusIndex == amfNameFocusCancel)
	actionLabel := "Create"
	if d.IsRename {
		actionLabel = "Rename"
	}
	confirmBtn := d.renderAMFButton(actionLabel, d.focusIndex == amfNameFocusConfirm)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)

	// Assemble
	content := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		nameLabel,
		nameField, "",
		buttons,
	)

	return d.borderStyle.Width(contentWidth).Render(content)
}

func (d *AdminMediaFolderNameDialogModel) renderAMFButton(text string, focused bool) string {
	style := d.buttonStyle
	if focused {
		style = style.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}
	return style.Render(buttonLabel(text, focused))
}

// AdminMediaFolderNameDialogCancelMsg signals the admin folder name dialog was cancelled.
type AdminMediaFolderNameDialogCancelMsg struct{}

// =============================================================================
// MOVE ADMIN MEDIA TO FOLDER DIALOG
// =============================================================================

// MoveAdminMediaFolderDialogModel presents a list of folders to move admin media into.
type MoveAdminMediaFolderDialogModel struct {
	dialogStyles

	Title        string
	Width        int
	AdminMediaID types.AdminMediaID
	Label        string // display name of the media being moved

	// Folder options: index 0 = "(root / unfiled)", then folders sorted alphabetically
	Options  []adminFolderOption
	Selected int

	focusIndex int // 0=list, 1=cancel, 2=move
}

type adminFolderOption struct {
	Label    string
	FolderID types.NullableAdminMediaFolderID
	Depth    int
}

const (
	amvFocusList    = 0
	amvFocusCancel  = 1
	amvFocusConfirm = 2
	amvMaxFocus     = 2
)

// NewMoveAdminMediaFolderDialog creates a dialog for selecting a destination admin folder.
func NewMoveAdminMediaFolderDialog(adminMediaID types.AdminMediaID, label string, folders []db.AdminMediaFolder) MoveAdminMediaFolderDialogModel {
	// Build flattened, indented folder list.
	options := []adminFolderOption{
		{Label: "(root / unfiled)", FolderID: types.NullableAdminMediaFolderID{}},
	}
	options = append(options, buildAdminFolderOptions(folders)...)

	return MoveAdminMediaFolderDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        fmt.Sprintf("Move '%s' to folder", label),
		Width:        55,
		AdminMediaID: adminMediaID,
		Label:        label,
		Options:      options,
		Selected:     0,
		focusIndex:   amvFocusList,
	}
}

// buildAdminFolderOptions creates a flat list of admin folder options with indentation based on nesting.
func buildAdminFolderOptions(folders []db.AdminMediaFolder) []adminFolderOption {
	// Index folders by ID
	childrenOf := make(map[types.AdminMediaFolderID][]db.AdminMediaFolder)
	var roots []db.AdminMediaFolder

	for _, f := range folders {
		if f.ParentID.Valid && !f.ParentID.ID.IsZero() {
			childrenOf[f.ParentID.ID] = append(childrenOf[f.ParentID.ID], f)
		} else {
			roots = append(roots, f)
		}
	}

	// Sort roots alphabetically
	sortAdminFolderSlice(roots)

	var result []adminFolderOption
	var walk func(fs []db.AdminMediaFolder, depth int)
	walk = func(fs []db.AdminMediaFolder, depth int) {
		for _, f := range fs {
			result = append(result, adminFolderOption{
				Label:    f.Name,
				FolderID: types.NullableAdminMediaFolderID{ID: f.AdminFolderID, Valid: true},
				Depth:    depth,
			})
			children := childrenOf[f.AdminFolderID]
			sortAdminFolderSlice(children)
			walk(children, depth+1)
		}
	}
	walk(roots, 0)
	return result
}

func sortAdminFolderSlice(fs []db.AdminMediaFolder) {
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && strings.ToLower(fs[j].Name) < strings.ToLower(fs[j-1].Name); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}

// Update handles user input for the move admin media dialog.
func (d *MoveAdminMediaFolderDialogModel) Update(msg tea.Msg) (MoveAdminMediaFolderDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			d.focusIndex++
			if d.focusIndex > amvMaxFocus {
				d.focusIndex = amvFocusList
			}
			return *d, nil
		case "shift+tab":
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = amvMaxFocus
			}
			return *d, nil
		case "up", "k":
			if d.focusIndex == amvFocusList && d.Selected > 0 {
				d.Selected--
			}
			return *d, nil
		case "down", "j":
			if d.focusIndex == amvFocusList && d.Selected < len(d.Options)-1 {
				d.Selected++
			}
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return MoveAdminMediaFolderDialogCancelMsg{} }
		case "enter":
			if d.focusIndex == amvFocusCancel {
				return *d, func() tea.Msg { return MoveAdminMediaFolderDialogCancelMsg{} }
			}
			if d.focusIndex == amvFocusConfirm || d.focusIndex == amvFocusList {
				adminMediaID := d.AdminMediaID
				folderID := d.Options[d.Selected].FolderID
				return *d, func() tea.Msg {
					return MoveAdminMediaToFolderRequestMsg{
						AdminMediaID: adminMediaID,
						FolderID:     folderID,
					}
				}
			}
			return *d, nil
		}
	}
	return *d, nil
}

// OverlayUpdate implements ModalOverlay for MoveAdminMediaFolderDialogModel.
func (d *MoveAdminMediaFolderDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for MoveAdminMediaFolderDialogModel.
func (d *MoveAdminMediaFolderDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the move admin media dialog with folder list.
func (d *MoveAdminMediaFolderDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	header := d.titleStyle.Render(d.Title)

	// Folder list
	var listLines []string
	accentStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	for i, opt := range d.Options {
		indent := strings.Repeat("  ", opt.Depth)
		cursor := "  "
		if d.Selected == i {
			cursor = ">"
			if d.focusIndex == amvFocusList {
				cursor = accentStyle.Render(">")
			}
		}
		line := fmt.Sprintf(" %s %s%s", cursor, indent, opt.Label)
		if d.Selected == i && d.focusIndex == amvFocusList {
			line = accentStyle.Render(line)
		}
		listLines = append(listLines, line)
	}

	// Limit visible lines to fit in terminal
	maxLines := windowHeight - 14
	if maxLines < 5 {
		maxLines = 5
	}
	if len(listLines) > maxLines {
		// Scroll to keep selected visible
		start := d.Selected - maxLines/2
		if start < 0 {
			start = 0
		}
		end := start + maxLines
		if end > len(listLines) {
			end = len(listLines)
			start = end - maxLines
			if start < 0 {
				start = 0
			}
		}
		listLines = listLines[start:end]
	}

	listContent := lipgloss.NewStyle().Width(innerW).Render(strings.Join(listLines, "\n"))

	// Buttons
	cancelBtn := d.renderAMVButton("Cancel", d.focusIndex == amvFocusCancel)
	moveBtn := d.renderAMVButton("Move", d.focusIndex == amvFocusConfirm)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", moveBtn)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		listContent, "",
		buttons,
	)

	return d.borderStyle.Width(contentWidth).Render(content)
}

func (d *MoveAdminMediaFolderDialogModel) renderAMVButton(text string, focused bool) string {
	style := d.buttonStyle
	if focused {
		style = style.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}
	return style.Render(buttonLabel(text, focused))
}

// MoveAdminMediaFolderDialogCancelMsg signals the move dialog was cancelled.
type MoveAdminMediaFolderDialogCancelMsg struct{}
