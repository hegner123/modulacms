package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/tui"
)

// DialogAction represents the type of action this dialog performs
type DialogAction string

const (
	DIALOGGENERIC       DialogAction = "generic"
	DIALOGDELETE        DialogAction = "delete"
	DIALOGACTIONCONFIRM DialogAction = "action_confirm"
	DIALOGINITCONTENT   DialogAction = "init_content"
	DIALOGQUITCONFIRM   DialogAction = "quit_confirm"
	DIALOGDELETECONTENT DialogAction = "delete_content"
)

// DialogModel represents a dialog that can be rendered on top of other content
type DialogModel struct {
	Title       string
	Message     string
	Width       int
	Height      int
	OkText      string
	CancelText  string
	ShowCancel  bool
	ReadyOK     bool
	Action      DialogAction
	ActionIndex int // Stores the pending action index for DIALOGACTIONCONFIRM
	help        help.Model
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	textStyle   lipgloss.Style
	buttonStyle lipgloss.Style
	focusIndex  int
}

// NewDialog creates a new dialog model with the given parameters
func NewDialog(title, message string, showCancel bool, action DialogAction) DialogModel {
	h := help.New()
	h.ShowAll = false

	return DialogModel{
		Title:       title,
		Message:     message,
		Width:       50,
		Height:      10,
		OkText:      "OK",
		CancelText:  "Cancel",
		ShowCancel:  showCancel,
		ReadyOK:     false,
		Action:      action,
		help:        h,
		borderStyle: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent),
		textStyle:   lipgloss.NewStyle().PaddingTop(1).PaddingBottom(1),
		buttonStyle: lipgloss.NewStyle().Padding(0, 2),
		focusIndex:  0,
	}
}

// SetSize sets the dialog size
func (d *DialogModel) SetSize(width, height int) {
	d.Width = width
	d.Height = height
}

// SetButtons sets the dialog button text
func (d *DialogModel) SetButtons(okText, cancelText string) {
	d.OkText = okText
	d.CancelText = cancelText
}

// Update handles user input for the dialog
func (d *DialogModel) Update(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch d.Action {
	case DIALOGDELETE, DIALOGACTIONCONFIRM, DIALOGINITCONTENT, DIALOGQUITCONFIRM, DIALOGDELETECONTENT:
		return d.ToggleControls(msg)
	case DIALOGGENERIC:
		// Generic dialog dismisses on enter or esc
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter", "esc":
				return *d, func() tea.Msg { return DialogCancelMsg{} }
			}
		}
		return *d, nil
	default:
		return *d, nil
	}
}

func (d *DialogModel) ToggleControls(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "h", "l", "j", "k", "left", "right", "up", "down":
			if d.ShowCancel {
				d.focusIndex = (d.focusIndex + 1) % 2
			}
			return *d, nil
		case "enter":
			if d.focusIndex == 0 {
				return *d, func() tea.Msg { return DialogAcceptMsg{Action: d.Action} }
			}
			return *d, func() tea.Msg { return DialogCancelMsg{} }
		case "esc":
			return *d, func() tea.Msg { return DialogCancelMsg{} }
		}
	}
	return *d, nil

}

// Render renders the dialog
func (d DialogModel) Render(windowWidth, windowHeight int) string {
	// Calculate position
	contentWidth := d.Width

	// Build dialog content
	titleText := d.titleStyle.Render(d.Title)
	messageText := d.textStyle.Render(d.Message)

	// Build buttons
	okButton := d.buttonStyle
	cancelButton := d.buttonStyle

	if d.focusIndex == 0 {
		okButton = okButton.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	} else {
		cancelButton = cancelButton.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}

	okButtonText := okButton.Render(d.OkText)
	cancelButtonText := cancelButton.Render(d.CancelText)

	// Position buttons
	var buttonBar string
	if d.ShowCancel {
		buttonBar = lipgloss.JoinHorizontal(lipgloss.Center, okButtonText, "  ", cancelButtonText)
	} else {
		buttonBar = lipgloss.JoinHorizontal(lipgloss.Center, okButtonText)
	}

	// Build the dialog box
	content := strings.Join([]string{
		titleText,
		"",
		messageText,
		"",
		buttonBar,
	}, "\n")

	// Apply border
	dialogBox := d.borderStyle.Width(contentWidth).Render(content)

	return dialogBox
}

// DialogOverlay positions a dialog over existing content using layer compositing.
func DialogOverlay(content string, dialog DialogModel, width, height int) string {
	dialogContent := dialog.Render(width, height)
	dialogW := lipgloss.Width(dialogContent)
	dialogH := lipgloss.Height(dialogContent)

	x := (width - dialogW) / 2
	y := (height - dialogH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return tui.Composite(content, tui.Overlay{
		Content: dialogContent,
		X:       x,
		Y:       y,
		Width:   dialogW,
		Height:  dialogH,
	})
}

// Dialog-related messages
type DialogAcceptMsg struct {
	Action DialogAction
}
type DialogCancelMsg struct{}
type DialogReadyOK struct{}
type ShowDialogMsg struct {
	Title      string
	Message    string
	ShowCancel bool
}

// HandleShowDialog creates a command to show a dialog
func HandleShowDialog(title, message string, showCancel bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDialogMsg{
			Title:      title,
			Message:    message,
			ShowCancel: showCancel,
		}
	}
}

// ShowQuitConfirmDialogMsg triggers showing a quit confirmation dialog
type ShowQuitConfirmDialogMsg struct{}

// ShowQuitConfirmDialogCmd creates a command to show a quit confirmation dialog
func ShowQuitConfirmDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return ShowQuitConfirmDialogMsg{}
	}
}

// ShowDeleteContentDialogMsg triggers showing a delete content confirmation dialog
type ShowDeleteContentDialogMsg struct {
	ContentID   string
	ContentName string
	HasChildren bool
}

// ShowDeleteContentDialogCmd creates a command to show a delete content confirmation dialog
func ShowDeleteContentDialogCmd(contentID, contentName string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteContentDialogMsg{
			ContentID:   contentID,
			ContentName: contentName,
			HasChildren: hasChildren,
		}
	}
}

// DeleteContentRequestMsg triggers content deletion
type DeleteContentRequestMsg struct {
	ContentID string
	RouteID   string
}

// DeleteContentCmd creates a command to delete content
func DeleteContentCmd(contentID, routeID string) tea.Cmd {
	return func() tea.Msg {
		return DeleteContentRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// ContentDeletedMsg is sent after content is successfully deleted
type ContentDeletedMsg struct {
	ContentID string
	RouteID   string
}
