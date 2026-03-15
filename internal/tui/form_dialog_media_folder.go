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
// MEDIA FOLDER NAME DIALOG (CREATE / RENAME)
// =============================================================================

// MediaFolderNameDialogModel is a simple single-input dialog for creating or renaming folders.
type MediaFolderNameDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	IsRename bool // true for rename, false for create

	NameInput textinput.Model

	// Context for the operation
	FolderID types.MediaFolderID         // set for rename only
	ParentID types.NullableMediaFolderID // set for create only

	focusIndex int // 0=input, 1=cancel, 2=confirm
}

const (
	mfNameFocusInput   = 0
	mfNameFocusCancel  = 1
	mfNameFocusConfirm = 2
	mfNameMaxFocus     = 2
)

// NewCreateFolderDialog creates a dialog for creating a new media folder.
func NewCreateFolderDialog(parentID types.NullableMediaFolderID) MediaFolderNameDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "folder name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.Focus()

	return MediaFolderNameDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        "New Folder",
		Width:        50,
		IsRename:     false,
		NameInput:    nameInput,
		ParentID:     parentID,
		focusIndex:   mfNameFocusInput,
	}
}

// NewRenameFolderDialog creates a dialog for renaming an existing media folder.
func NewRenameFolderDialog(folderID types.MediaFolderID, currentName string) MediaFolderNameDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "new name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.SetValue(currentName)
	nameInput.Focus()

	return MediaFolderNameDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        "Rename Folder",
		Width:        50,
		IsRename:     true,
		NameInput:    nameInput,
		FolderID:     folderID,
		focusIndex:   mfNameFocusInput,
	}
}

// Update handles user input for the folder name dialog.
func (d *MediaFolderNameDialogModel) Update(msg tea.Msg) (MediaFolderNameDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.focusIndex++
			if d.focusIndex > mfNameMaxFocus {
				d.focusIndex = mfNameFocusInput
			}
			d.updateFocus()
			return *d, nil
		case "shift+tab", "up":
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = mfNameMaxFocus
			}
			d.updateFocus()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return MediaFolderNameDialogCancelMsg{} }
		case "enter":
			if d.focusIndex == mfNameFocusCancel {
				return *d, func() tea.Msg { return MediaFolderNameDialogCancelMsg{} }
			}
			if d.focusIndex == mfNameFocusConfirm || d.focusIndex == mfNameFocusInput {
				name := strings.TrimSpace(d.NameInput.Value())
				if name == "" {
					return *d, nil // no-op on empty name
				}
				if d.IsRename {
					folderID := d.FolderID
					return *d, func() tea.Msg {
						return RenameMediaFolderRequestMsg{
							FolderID: folderID,
							NewName:  name,
						}
					}
				}
				parentID := d.ParentID
				return *d, func() tea.Msg {
					return CreateMediaFolderRequestMsg{
						Name:     name,
						ParentID: parentID,
					}
				}
			}
			return *d, nil
		default:
			if d.focusIndex == mfNameFocusInput {
				var cmd tea.Cmd
				d.NameInput, cmd = d.NameInput.Update(msg)
				return *d, cmd
			}
		}
	}
	return *d, nil
}

func (d *MediaFolderNameDialogModel) updateFocus() {
	if d.focusIndex == mfNameFocusInput {
		d.NameInput.Focus()
	} else {
		d.NameInput.Blur()
	}
}

// OverlayUpdate implements ModalOverlay for MediaFolderNameDialogModel.
func (d *MediaFolderNameDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayTick implements OverlayTicker for cursor blinking.
func (d *MediaFolderNameDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	if d.focusIndex == mfNameFocusInput {
		var cmd tea.Cmd
		d.NameInput, cmd = d.NameInput.Update(msg)
		return d, cmd
	}
	return d, nil
}

// OverlayView implements ModalOverlay for MediaFolderNameDialogModel.
func (d *MediaFolderNameDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the folder name dialog.
func (d *MediaFolderNameDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	// Header
	header := d.titleStyle.Render(d.Title)

	// Name input field
	nameLabel := d.labelStyle.Render("Name")
	nameInputStyle := d.inputStyle
	if d.focusIndex == mfNameFocusInput {
		nameInputStyle = d.focusedInputStyle
	}
	nameField := nameInputStyle.Width(innerW).Render(d.NameInput.View())

	// Buttons
	cancelBtn := d.renderButton("Cancel", d.focusIndex == mfNameFocusCancel)
	actionLabel := "Create"
	if d.IsRename {
		actionLabel = "Rename"
	}
	confirmBtn := d.renderButton(actionLabel, d.focusIndex == mfNameFocusConfirm)
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

func (d *MediaFolderNameDialogModel) renderButton(text string, focused bool) string {
	style := d.buttonStyle
	if focused {
		style = style.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}
	return style.Render(buttonLabel(text, focused))
}

// MediaFolderNameDialogCancelMsg signals the folder name dialog was cancelled.
type MediaFolderNameDialogCancelMsg struct{}

// =============================================================================
// MOVE MEDIA TO FOLDER DIALOG
// =============================================================================

// MoveMediaFolderDialogModel presents a list of folders to move media into.
type MoveMediaFolderDialogModel struct {
	dialogStyles

	Title   string
	Width   int
	MediaID types.MediaID
	Label   string // display name of the media being moved

	// Folder options: index 0 = "(root / unfiled)", then folders sorted alphabetically
	Options  []folderOption
	Selected int

	focusIndex int // 0=list, 1=cancel, 2=move
}

type folderOption struct {
	Label    string
	FolderID types.NullableMediaFolderID
	Depth    int
}

const (
	mvFocusList    = 0
	mvFocusCancel  = 1
	mvFocusConfirm = 2
	mvMaxFocus     = 2
)

// NewMoveMediaFolderDialog creates a dialog for selecting a destination folder.
func NewMoveMediaFolderDialog(mediaID types.MediaID, label string, folders []db.MediaFolder) MoveMediaFolderDialogModel {
	// Build flattened, indented folder list.
	options := []folderOption{
		{Label: "(root / unfiled)", FolderID: types.NullableMediaFolderID{}},
	}
	// Build a tree structure to show folders indented
	options = append(options, buildFolderOptions(folders)...)

	return MoveMediaFolderDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        fmt.Sprintf("Move '%s' to folder", label),
		Width:        55,
		MediaID:      mediaID,
		Label:        label,
		Options:      options,
		Selected:     0,
		focusIndex:   mvFocusList,
	}
}

// buildFolderOptions creates a flat list of folder options with indentation based on nesting.
func buildFolderOptions(folders []db.MediaFolder) []folderOption {
	// Index folders by ID
	byID := make(map[types.MediaFolderID]db.MediaFolder, len(folders))
	childrenOf := make(map[types.MediaFolderID][]db.MediaFolder)
	var roots []db.MediaFolder

	for _, f := range folders {
		byID[f.FolderID] = f
		if f.ParentID.Valid && !f.ParentID.ID.IsZero() {
			childrenOf[f.ParentID.ID] = append(childrenOf[f.ParentID.ID], f)
		} else {
			roots = append(roots, f)
		}
	}

	// Sort roots alphabetically
	sortFolderSlice(roots)

	var result []folderOption
	var walk func(fs []db.MediaFolder, depth int)
	walk = func(fs []db.MediaFolder, depth int) {
		for _, f := range fs {
			result = append(result, folderOption{
				Label:    f.Name,
				FolderID: types.NullableMediaFolderID{ID: f.FolderID, Valid: true},
				Depth:    depth,
			})
			children := childrenOf[f.FolderID]
			sortFolderSlice(children)
			walk(children, depth+1)
		}
	}
	walk(roots, 0)
	return result
}

func sortFolderSlice(fs []db.MediaFolder) {
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && strings.ToLower(fs[j].Name) < strings.ToLower(fs[j-1].Name); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}

// Update handles user input for the move media dialog.
func (d *MoveMediaFolderDialogModel) Update(msg tea.Msg) (MoveMediaFolderDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			d.focusIndex++
			if d.focusIndex > mvMaxFocus {
				d.focusIndex = mvFocusList
			}
			return *d, nil
		case "shift+tab":
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = mvMaxFocus
			}
			return *d, nil
		case "up", "k":
			if d.focusIndex == mvFocusList && d.Selected > 0 {
				d.Selected--
			}
			return *d, nil
		case "down", "j":
			if d.focusIndex == mvFocusList && d.Selected < len(d.Options)-1 {
				d.Selected++
			}
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return MoveMediaFolderDialogCancelMsg{} }
		case "enter":
			if d.focusIndex == mvFocusCancel {
				return *d, func() tea.Msg { return MoveMediaFolderDialogCancelMsg{} }
			}
			if d.focusIndex == mvFocusConfirm || d.focusIndex == mvFocusList {
				mediaID := d.MediaID
				folderID := d.Options[d.Selected].FolderID
				return *d, func() tea.Msg {
					return MoveMediaToFolderRequestMsg{
						MediaID:  mediaID,
						FolderID: folderID,
					}
				}
			}
			return *d, nil
		}
	}
	return *d, nil
}

// OverlayUpdate implements ModalOverlay for MoveMediaFolderDialogModel.
func (d *MoveMediaFolderDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for MoveMediaFolderDialogModel.
func (d *MoveMediaFolderDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the move media dialog with folder list.
func (d *MoveMediaFolderDialogModel) Render(windowWidth, windowHeight int) string {
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
			if d.focusIndex == mvFocusList {
				cursor = accentStyle.Render(">")
			}
		}
		line := fmt.Sprintf(" %s %s%s", cursor, indent, opt.Label)
		if d.Selected == i && d.focusIndex == mvFocusList {
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
	cancelBtn := d.renderMoveButton("Cancel", d.focusIndex == mvFocusCancel)
	moveBtn := d.renderMoveButton("Move", d.focusIndex == mvFocusConfirm)
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", moveBtn)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		listContent, "",
		buttons,
	)

	return d.borderStyle.Width(contentWidth).Render(content)
}

func (d *MoveMediaFolderDialogModel) renderMoveButton(text string, focused bool) string {
	style := d.buttonStyle
	if focused {
		style = style.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}
	return style.Render(buttonLabel(text, focused))
}

// MoveMediaFolderDialogCancelMsg signals the move dialog was cancelled.
type MoveMediaFolderDialogCancelMsg struct{}
