package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
)

// =============================================================================
// TOKEN FORM DIALOG
// =============================================================================

// TokenFormDialogModel represents a form dialog for creating API tokens.
// Tokens only have a type selector — the token value is generated server-side.
type TokenFormDialogModel struct {
	dialogStyles

	Title string
	Width int

	TypeInput textinput.Model

	// Focus indices: 0=Type, 1=Cancel, 2=Confirm
	focusIndex int
}

const (
	tokenFocusType    = 0
	tokenFocusCancel  = 1
	tokenFocusConfirm = 2
	tokenMaxFocus     = 2
)

// NewTokenFormDialog creates a token form dialog for creating a new API token.
func NewTokenFormDialog(title string) TokenFormDialogModel {
	typeInput := textinput.New()
	typeInput.Placeholder = "api"
	typeInput.CharLimit = 64
	typeInput.SetWidth(40)
	typeInput.Focus()

	return TokenFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        50,
		TypeInput:    typeInput,
		focusIndex:   tokenFocusType,
	}
}

// Update handles user input for the token form dialog.
func (d *TokenFormDialogModel) Update(msg tea.Msg) (TokenFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.tokenFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.tokenFocusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == tokenFocusConfirm {
				d.focusIndex = tokenFocusCancel
				return *d, nil
			}
		case "right":
			if d.focusIndex == tokenFocusCancel {
				d.focusIndex = tokenFocusConfirm
				return *d, nil
			}
		case "enter":
			if d.focusIndex == tokenFocusConfirm {
				tokenType := d.TypeInput.Value()
				if tokenType == "" {
					tokenType = "api"
				}
				return *d, func() tea.Msg {
					return TokenFormDialogAcceptMsg{
						TokenType: tokenType,
					}
				}
			}
			if d.focusIndex == tokenFocusCancel {
				return *d, func() tea.Msg { return TokenFormDialogCancelMsg{} }
			}
			d.tokenFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return TokenFormDialogCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	if d.focusIndex == tokenFocusType {
		d.TypeInput, cmd = d.TypeInput.Update(msg)
	}
	return *d, cmd
}

func (d *TokenFormDialogModel) tokenFocusNext() {
	d.focusIndex = (d.focusIndex + 1) % (tokenMaxFocus + 1)
	d.tokenUpdateFocus()
}

func (d *TokenFormDialogModel) tokenFocusPrev() {
	d.focusIndex = (d.focusIndex + tokenMaxFocus) % (tokenMaxFocus + 1)
	d.tokenUpdateFocus()
}

func (d *TokenFormDialogModel) tokenUpdateFocus() {
	d.TypeInput.Blur()
	if d.focusIndex == tokenFocusType {
		d.TypeInput.Focus()
	}
}

// OverlayUpdate implements ModalOverlay for TokenFormDialogModel.
func (d *TokenFormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayTick forwards non-key messages to the focused input.
func (d *TokenFormDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	var cmd tea.Cmd
	if d.focusIndex == tokenFocusType {
		d.TypeInput, cmd = d.TypeInput.Update(msg)
	}
	return d, cmd
}

// OverlayView implements ModalOverlay for TokenFormDialogModel.
func (d *TokenFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the token form dialog.
func (d TokenFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	content.WriteString(d.labelStyle.Render("Token Type"))
	content.WriteString("\n")
	if d.focusIndex == tokenFocusType {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.TypeInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.TypeInput.View()))
	}
	content.WriteString("\n\n")

	content.WriteString(lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Render("A random API token will be generated.\nThe token value is shown only once."))
	content.WriteString("\n\n")

	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == tokenFocusCancel {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == tokenFocusConfirm {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == tokenFocusCancel))
	confirmBtn := confirmStyle.Render(buttonLabel("Create", d.focusIndex == tokenFocusConfirm))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// =============================================================================
// TOKEN FORM DIALOG MESSAGES
// =============================================================================

// TokenFormDialogAcceptMsg is sent when the token form dialog is confirmed.
type TokenFormDialogAcceptMsg struct {
	TokenType string
}

// TokenFormDialogCancelMsg is sent when the token form dialog is cancelled.
type TokenFormDialogCancelMsg struct{}

// ShowTokenFormDialogMsg triggers showing a token creation form dialog.
type ShowTokenFormDialogMsg struct {
	Title string
}
