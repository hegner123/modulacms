package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// =============================================================================
// WEBHOOK FORM DIALOG
// =============================================================================

// WebhookFormDialogModel represents a form dialog for webhook CRUD operations.
type WebhookFormDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	Action   FormDialogAction
	EntityID string // WebhookID for edit operations

	NameInput   textinput.Model
	URLInput    textinput.Model
	SecretInput textinput.Model
	EventsInput textinput.Model // comma-separated event types

	// Active toggle carousel
	ActiveOptions []string // ["Yes", "No"]
	ActiveIndex   int      // 0=Yes, 1=No

	// Focus indices:
	// 0=Name, 1=URL, 2=Secret, 3=Events, 4=Active, 5=Cancel, 6=Confirm
	focusIndex int
}

const (
	webhookFocusName    = 0
	webhookFocusURL     = 1
	webhookFocusSecret  = 2
	webhookFocusEvents  = 3
	webhookFocusActive  = 4
	webhookFocusCancel  = 5
	webhookFocusConfirm = 6
	webhookMaxFocus     = 6
)

// NewWebhookFormDialog creates a webhook form dialog for creating a new webhook.
func NewWebhookFormDialog(title string) WebhookFormDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Webhook name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.Focus()

	urlInput := textinput.New()
	urlInput.Placeholder = "https://example.com/webhook"
	urlInput.CharLimit = 512
	urlInput.SetWidth(40)

	secretInput := textinput.New()
	secretInput.Placeholder = "signing secret (optional)"
	secretInput.CharLimit = 256
	secretInput.SetWidth(40)

	eventsInput := textinput.New()
	eventsInput.Placeholder = "content.created, content.updated"
	eventsInput.CharLimit = 512
	eventsInput.SetWidth(40)

	return WebhookFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGCREATEWEBHOOK,
		NameInput:     nameInput,
		URLInput:      urlInput,
		SecretInput:   secretInput,
		EventsInput:   eventsInput,
		ActiveOptions: []string{"Yes", "No"},
		ActiveIndex:   0, // default to active
		focusIndex:    webhookFocusName,
	}
}

// NewEditWebhookFormDialog creates a webhook form dialog pre-populated for editing.
func NewEditWebhookFormDialog(title string, webhook db.Webhook) WebhookFormDialogModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Webhook name"
	nameInput.CharLimit = 128
	nameInput.SetWidth(40)
	nameInput.SetValue(webhook.Name)
	nameInput.Focus()

	urlInput := textinput.New()
	urlInput.Placeholder = "https://example.com/webhook"
	urlInput.CharLimit = 512
	urlInput.SetWidth(40)
	urlInput.SetValue(webhook.URL)

	secretInput := textinput.New()
	secretInput.Placeholder = "signing secret (optional)"
	secretInput.CharLimit = 256
	secretInput.SetWidth(40)
	secretInput.SetValue(webhook.Secret)

	eventsInput := textinput.New()
	eventsInput.Placeholder = "content.created, content.updated"
	eventsInput.CharLimit = 512
	eventsInput.SetWidth(40)
	eventsInput.SetValue(strings.Join(webhook.Events, ", "))

	activeIndex := 0
	if !webhook.IsActive {
		activeIndex = 1
	}

	return WebhookFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGEDITWEBHOOK,
		EntityID:      string(webhook.WebhookID),
		NameInput:     nameInput,
		URLInput:      urlInput,
		SecretInput:   secretInput,
		EventsInput:   eventsInput,
		ActiveOptions: []string{"Yes", "No"},
		ActiveIndex:   activeIndex,
		focusIndex:    webhookFocusName,
	}
}

// Update handles user input for the webhook form dialog.
func (d *WebhookFormDialogModel) Update(msg tea.Msg) (WebhookFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.webhookFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.webhookFocusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == webhookFocusActive && len(d.ActiveOptions) > 0 {
				if d.ActiveIndex > 0 {
					d.ActiveIndex--
				}
				return *d, nil
			}
			if d.focusIndex == webhookFocusConfirm {
				d.focusIndex = webhookFocusCancel
				return *d, nil
			}
		case "right":
			if d.focusIndex == webhookFocusActive && len(d.ActiveOptions) > 0 {
				if d.ActiveIndex < len(d.ActiveOptions)-1 {
					d.ActiveIndex++
				}
				return *d, nil
			}
			if d.focusIndex == webhookFocusCancel {
				d.focusIndex = webhookFocusConfirm
				return *d, nil
			}
		case "enter":
			if d.focusIndex == webhookFocusConfirm {
				isActive := d.ActiveIndex == 0
				return *d, func() tea.Msg {
					return WebhookFormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Name:     d.NameInput.Value(),
						URL:      d.URLInput.Value(),
						Secret:   d.SecretInput.Value(),
						Events:   d.EventsInput.Value(),
						IsActive: isActive,
					}
				}
			}
			if d.focusIndex == webhookFocusCancel {
				return *d, func() tea.Msg { return WebhookFormDialogCancelMsg{} }
			}
			// On text fields, move to next
			d.webhookFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return WebhookFormDialogCancelMsg{} }
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	switch d.focusIndex {
	case webhookFocusName:
		d.NameInput, cmd = d.NameInput.Update(msg)
	case webhookFocusURL:
		d.URLInput, cmd = d.URLInput.Update(msg)
	case webhookFocusSecret:
		d.SecretInput, cmd = d.SecretInput.Update(msg)
	case webhookFocusEvents:
		d.EventsInput, cmd = d.EventsInput.Update(msg)
	}
	return *d, cmd
}

// webhookFocusNext advances focus to the next focusable element, wrapping at the end.
func (d *WebhookFormDialogModel) webhookFocusNext() {
	d.focusIndex = (d.focusIndex + 1) % (webhookMaxFocus + 1)
	d.webhookUpdateFocus()
}

// webhookFocusPrev moves focus to the previous focusable element, wrapping at the start.
func (d *WebhookFormDialogModel) webhookFocusPrev() {
	d.focusIndex = (d.focusIndex + webhookMaxFocus) % (webhookMaxFocus + 1)
	d.webhookUpdateFocus()
}

// webhookUpdateFocus applies focus styling to the currently focused field.
func (d *WebhookFormDialogModel) webhookUpdateFocus() {
	d.NameInput.Blur()
	d.URLInput.Blur()
	d.SecretInput.Blur()
	d.EventsInput.Blur()
	switch d.focusIndex {
	case webhookFocusName:
		d.NameInput.Focus()
	case webhookFocusURL:
		d.URLInput.Focus()
	case webhookFocusSecret:
		d.SecretInput.Focus()
	case webhookFocusEvents:
		d.EventsInput.Focus()
	}
}

// OverlayUpdate implements ModalOverlay for WebhookFormDialogModel.
func (d *WebhookFormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayTick forwards non-key messages (cursor blink, etc.) to the
// focused text input so it can animate and re-render correctly.
func (d *WebhookFormDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	var cmd tea.Cmd
	switch d.focusIndex {
	case webhookFocusName:
		d.NameInput, cmd = d.NameInput.Update(msg)
	case webhookFocusURL:
		d.URLInput, cmd = d.URLInput.Update(msg)
	case webhookFocusSecret:
		d.SecretInput, cmd = d.SecretInput.Update(msg)
	case webhookFocusEvents:
		d.EventsInput, cmd = d.EventsInput.Update(msg)
	}
	return d, cmd
}

// OverlayView implements ModalOverlay for WebhookFormDialogModel.
func (d *WebhookFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the webhook form dialog.
func (d WebhookFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Name field
	content.WriteString(d.labelStyle.Render("Name"))
	content.WriteString("\n")
	if d.focusIndex == webhookFocusName {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.NameInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.NameInput.View()))
	}
	content.WriteString("\n")

	// URL field
	content.WriteString(d.labelStyle.Render("URL"))
	content.WriteString("\n")
	if d.focusIndex == webhookFocusURL {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.URLInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.URLInput.View()))
	}
	content.WriteString("\n")

	// Secret field
	content.WriteString(d.labelStyle.Render("Secret"))
	content.WriteString("\n")
	if d.focusIndex == webhookFocusSecret {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.SecretInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.SecretInput.View()))
	}
	content.WriteString("\n")

	// Events field
	content.WriteString(d.labelStyle.Render("Events (comma-separated)"))
	content.WriteString("\n")
	if d.focusIndex == webhookFocusEvents {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.EventsInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.EventsInput.View()))
	}
	content.WriteString("\n")

	// Active carousel
	content.WriteString(d.labelStyle.Render("Active"))
	content.WriteString("\n")
	if len(d.ActiveOptions) > 0 {
		optLabel := d.ActiveOptions[d.ActiveIndex]
		if d.focusIndex == webhookFocusActive {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Primary).
				Background(config.DefaultStyle.Accent).
				Padding(0, 1).
				Render(string(rune(9664)) + " " + optLabel + " " + string(rune(9654)))
			content.WriteString(selector)
		} else {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Secondary).
				Padding(0, 1).
				Render("  " + optLabel + "  ")
			content.WriteString(selector)
		}
	}
	content.WriteString("\n\n")

	// Buttons
	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == webhookFocusCancel {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == webhookFocusConfirm {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == webhookFocusCancel))
	confirmBtn := confirmStyle.Render(buttonLabel("Save", d.focusIndex == webhookFocusConfirm))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// =============================================================================
// WEBHOOK FORM DIALOG MESSAGES
// =============================================================================

// WebhookFormDialogAcceptMsg is sent when the webhook form dialog is confirmed.
type WebhookFormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string
	Name     string
	URL      string
	Secret   string
	Events   string // comma-separated, parsed by handler
	IsActive bool
}

// WebhookFormDialogCancelMsg is sent when the webhook form dialog is cancelled.
type WebhookFormDialogCancelMsg struct{}

// ShowWebhookFormDialogMsg triggers showing a webhook creation form dialog.
type ShowWebhookFormDialogMsg struct {
	Title string
}

// ShowEditWebhookDialogMsg triggers showing a webhook edit form dialog.
type ShowEditWebhookDialogMsg struct {
	Webhook db.Webhook
}
