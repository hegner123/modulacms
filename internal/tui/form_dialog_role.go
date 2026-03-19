package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// RoleFormDialogModel represents a form dialog for role CRUD.
type RoleFormDialogModel struct {
	dialogStyles
	Title    string
	Width    int
	Action   FormDialogAction
	EntityID string // RoleID for edit

	LabelInput textinput.Model

	// 0=Label, 1=Cancel, 2=Confirm
	focusIndex int
}

const (
	roleFocusLabel   = 0
	roleFocusCancel  = 1
	roleFocusConfirm = 2
	roleMaxFocus     = 2
)

func NewRoleFormDialog(title string) RoleFormDialogModel {
	labelInput := textinput.New()
	labelInput.Placeholder = "role name"
	labelInput.CharLimit = 64
	labelInput.SetWidth(40)
	labelInput.Focus()

	return RoleFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        50,
		Action:       FORMDIALOGCREATEROLE,
		LabelInput:   labelInput,
		focusIndex:   roleFocusLabel,
	}
}

func NewEditRoleFormDialog(title string, role db.Roles) RoleFormDialogModel {
	labelInput := textinput.New()
	labelInput.Placeholder = "role name"
	labelInput.CharLimit = 64
	labelInput.SetWidth(40)
	labelInput.SetValue(role.Label)
	labelInput.Focus()

	return RoleFormDialogModel{
		dialogStyles: newDialogStyles(),
		Title:        title,
		Width:        50,
		Action:       FORMDIALOGEDITROLE,
		EntityID:     string(role.RoleID),
		LabelInput:   labelInput,
		focusIndex:   roleFocusLabel,
	}
}

func (d *RoleFormDialogModel) Update(msg tea.Msg) (RoleFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.focusIndex = (d.focusIndex + 1) % (roleMaxFocus + 1)
			d.updateRoleFocus()
			return *d, nil
		case "shift+tab", "up":
			d.focusIndex = (d.focusIndex + roleMaxFocus) % (roleMaxFocus + 1)
			d.updateRoleFocus()
			return *d, nil
		case "left":
			if d.focusIndex == roleFocusConfirm {
				d.focusIndex = roleFocusCancel
				return *d, nil
			}
		case "right":
			if d.focusIndex == roleFocusCancel {
				d.focusIndex = roleFocusConfirm
				return *d, nil
			}
		case "enter":
			if d.focusIndex == roleFocusConfirm {
				return *d, func() tea.Msg {
					return RoleFormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Label:    d.LabelInput.Value(),
					}
				}
			}
			if d.focusIndex == roleFocusCancel {
				return *d, func() tea.Msg { return RoleFormDialogCancelMsg{} }
			}
			d.focusIndex = (d.focusIndex + 1) % (roleMaxFocus + 1)
			d.updateRoleFocus()
			return *d, nil
		case "escape":
			return *d, func() tea.Msg { return RoleFormDialogCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	if d.focusIndex == roleFocusLabel {
		d.LabelInput, cmd = d.LabelInput.Update(msg)
	}
	return *d, cmd
}

func (d *RoleFormDialogModel) updateRoleFocus() {
	d.LabelInput.Blur()
	if d.focusIndex == roleFocusLabel {
		d.LabelInput.Focus()
	}
}

func (d *RoleFormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

func (d *RoleFormDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	var cmd tea.Cmd
	if d.focusIndex == roleFocusLabel {
		d.LabelInput, cmd = d.LabelInput.Update(msg)
	}
	return d, cmd
}

func (d *RoleFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

func (d RoleFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	content.WriteString(d.labelStyle.Render("Role Name"))
	content.WriteString("\n")
	if d.focusIndex == roleFocusLabel {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.LabelInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.LabelInput.View()))
	}
	content.WriteString("\n\n")

	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == roleFocusCancel {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == roleFocusConfirm {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == roleFocusCancel))
	confirmBtn := confirmStyle.Render(buttonLabel("Save", d.focusIndex == roleFocusConfirm))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// Messages

type RoleFormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string
	Label    string
}

type RoleFormDialogCancelMsg struct{}

type ShowRoleFormDialogMsg struct {
	Title string
}

type ShowEditRoleDialogMsg struct {
	Role db.Roles
}
