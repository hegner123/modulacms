package cli

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
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
	help        help.Model
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	textStyle   lipgloss.Style
	buttonStyle lipgloss.Style
	focusIndex  int
}

// NewDialog creates a new dialog model with the given parameters
func NewDialog(title, message string, showCancel bool) DialogModel {
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "h", "l":
			if d.ShowCancel {
				d.focusIndex = (d.focusIndex + 1) % 2
			}
			return *d, nil
		case "enter":
			if d.focusIndex == 0 {
				return *d, func() tea.Msg { return DialogAcceptMsg{} }
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

	// Apply border and position
	dialogBox := d.borderStyle.Width(contentWidth).Render(content)

	// Center the dialog on screen
	dialogBox = lipgloss.Place(
		windowWidth,
		windowHeight,
		lipgloss.Center,
		lipgloss.Center,
		dialogBox,
	)

	return dialogBox
}

// DialogOverlay positions a dialog over existing content
func DialogOverlay(content string, dialog DialogModel, width, height int) string {
	dialogContent := dialog.Render(width, height)

	// Place the dialog on top of the overlay
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialogContent,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")))
}

// Dialog-related messages
type DialogAcceptMsg struct{}
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
