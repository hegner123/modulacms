package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// =============================================================================
// USER FORM DIALOG
// =============================================================================

// UserFormDialogModel represents a form dialog for user CRUD operations.
type UserFormDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	Action   FormDialogAction
	EntityID string // UserID for edit operations

	UsernameInput textinput.Model
	NameInput     textinput.Model
	EmailInput    textinput.Model
	PasswordInput textinput.Model

	// Role selection carousel
	RoleOptions []RoleOption
	RoleIndex   int

	// Create mode focus: 0=username, 1=name, 2=email, 3=password, 4=role, 5=cancel, 6=confirm
	// Edit mode focus:   0=username, 1=name, 2=email, 3=role, 4=cancel, 5=confirm
	focusIndex int
}

// isCreateMode returns true if the dialog is for creating a new user.
func (d *UserFormDialogModel) isCreateMode() bool {
	return d.Action == FORMDIALOGCREATEUSER
}

// maxFocusIndex returns the maximum valid focus index.
func (d *UserFormDialogModel) maxFocusIndex() int {
	if d.isCreateMode() {
		return 6
	}
	return 5
}

// roleFocusIndex returns the focus index for the role carousel.
func (d *UserFormDialogModel) roleFocusIndex() int {
	if d.isCreateMode() {
		return 4
	}
	return 3
}

// cancelFocusIndex returns the focus index for the cancel button.
func (d *UserFormDialogModel) cancelFocusIndex() int {
	if d.isCreateMode() {
		return 5
	}
	return 4
}

// confirmFocusIndex returns the focus index for the confirm button.
func (d *UserFormDialogModel) confirmFocusIndex() int {
	if d.isCreateMode() {
		return 6
	}
	return 5
}

// NewUserFormDialog creates a user form dialog for creating a new user
func NewUserFormDialog(title string, roles []db.Roles) UserFormDialogModel {
	username := textinput.New()
	username.Placeholder = "username"
	username.CharLimit = 64
	username.Width = 40
	username.Focus()

	name := textinput.New()
	name.Placeholder = "Full Name"
	name.CharLimit = 128
	name.Width = 40

	email := textinput.New()
	email.Placeholder = "user@example.com"
	email.CharLimit = 128
	email.Width = 40

	password := textinput.New()
	password.Placeholder = "password"
	password.CharLimit = 72
	password.Width = 40
	password.EchoMode = textinput.EchoPassword

	roleOptions := make([]RoleOption, 0, len(roles))
	for _, r := range roles {
		roleOptions = append(roleOptions, RoleOption{
			Label: r.Label,
			Value: string(r.RoleID),
		})
	}

	return UserFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGCREATEUSER,
		UsernameInput: username,
		NameInput:     name,
		EmailInput:    email,
		PasswordInput: password,
		RoleOptions:   roleOptions,
		RoleIndex:     0,
		focusIndex:    0,
	}
}

// NewEditUserFormDialog creates a user form dialog pre-populated for editing
func NewEditUserFormDialog(title string, user db.UserWithRoleLabelRow, roles []db.Roles) UserFormDialogModel {
	username := textinput.New()
	username.Placeholder = "username"
	username.CharLimit = 64
	username.Width = 40
	username.SetValue(user.Username)
	username.Focus()

	name := textinput.New()
	name.Placeholder = "Full Name"
	name.CharLimit = 128
	name.Width = 40
	name.SetValue(user.Name)

	email := textinput.New()
	email.Placeholder = "user@example.com"
	email.CharLimit = 128
	email.Width = 40
	email.SetValue(user.Email.String())

	// Password input exists but is not shown in edit mode
	password := textinput.New()
	password.Placeholder = ""
	password.CharLimit = 72
	password.Width = 40
	password.EchoMode = textinput.EchoPassword

	roleOptions := make([]RoleOption, 0, len(roles))
	selectedRoleIndex := 0
	for i, r := range roles {
		roleOptions = append(roleOptions, RoleOption{
			Label: r.Label,
			Value: string(r.RoleID),
		})
		if string(r.RoleID) == user.Role {
			selectedRoleIndex = i
		}
	}

	return UserFormDialogModel{
		dialogStyles:  newDialogStyles(),
		Title:         title,
		Width:         60,
		Action:        FORMDIALOGEDITUSER,
		EntityID:      user.UserID.String(),
		UsernameInput: username,
		NameInput:     name,
		EmailInput:    email,
		PasswordInput: password,
		RoleOptions:   roleOptions,
		RoleIndex:     selectedRoleIndex,
		focusIndex:    0,
	}
}

// Update handles user input for the user form dialog
func (d *UserFormDialogModel) Update(msg tea.Msg) (UserFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			d.userFormFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.userFormFocusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == d.roleFocusIndex() && len(d.RoleOptions) > 0 {
				if d.RoleIndex > 0 {
					d.RoleIndex--
				}
				return *d, nil
			}
			if d.focusIndex == d.confirmFocusIndex() {
				d.focusIndex = d.cancelFocusIndex()
				return *d, nil
			}
		case "right":
			if d.focusIndex == d.roleFocusIndex() && len(d.RoleOptions) > 0 {
				if d.RoleIndex < len(d.RoleOptions)-1 {
					d.RoleIndex++
				}
				return *d, nil
			}
			if d.focusIndex == d.cancelFocusIndex() {
				d.focusIndex = d.confirmFocusIndex()
				return *d, nil
			}
		case "enter":
			if d.focusIndex == d.confirmFocusIndex() {
				roleValue := ""
				if len(d.RoleOptions) > 0 && d.RoleIndex < len(d.RoleOptions) {
					roleValue = d.RoleOptions[d.RoleIndex].Value
				}
				password := ""
				if d.isCreateMode() {
					password = d.PasswordInput.Value()
				}
				return *d, func() tea.Msg {
					return UserFormDialogAcceptMsg{
						Action:   d.Action,
						EntityID: d.EntityID,
						Username: d.UsernameInput.Value(),
						Name:     d.NameInput.Value(),
						Email:    d.EmailInput.Value(),
						Password: password,
						Role:     roleValue,
					}
				}
			}
			if d.focusIndex == d.cancelFocusIndex() {
				return *d, func() tea.Msg { return UserFormDialogCancelMsg{} }
			}
			// On text fields, move to next
			d.userFormFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return UserFormDialogCancelMsg{} }
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	switch d.focusIndex {
	case 0:
		d.UsernameInput, cmd = d.UsernameInput.Update(msg)
	case 1:
		d.NameInput, cmd = d.NameInput.Update(msg)
	case 2:
		d.EmailInput, cmd = d.EmailInput.Update(msg)
	default:
		if d.isCreateMode() && d.focusIndex == 3 {
			d.PasswordInput, cmd = d.PasswordInput.Update(msg)
		}
	}
	return *d, cmd
}

// userFormFocusNext advances focus to the next focusable element in the user form, wrapping at the end.
func (d *UserFormDialogModel) userFormFocusNext() {
	max := d.maxFocusIndex()
	d.focusIndex = (d.focusIndex + 1) % (max + 1)
	d.userFormUpdateFocus()
}

// userFormFocusPrev moves focus to the previous focusable element in the user form, wrapping at the start.
func (d *UserFormDialogModel) userFormFocusPrev() {
	max := d.maxFocusIndex()
	d.focusIndex = (d.focusIndex + max) % (max + 1)
	d.userFormUpdateFocus()
}

// userFormUpdateFocus applies focus styling to the currently focused field in the user form.
func (d *UserFormDialogModel) userFormUpdateFocus() {
	d.UsernameInput.Blur()
	d.NameInput.Blur()
	d.EmailInput.Blur()
	d.PasswordInput.Blur()
	switch d.focusIndex {
	case 0:
		d.UsernameInput.Focus()
	case 1:
		d.NameInput.Focus()
	case 2:
		d.EmailInput.Focus()
	default:
		if d.isCreateMode() && d.focusIndex == 3 {
			d.PasswordInput.Focus()
		}
	}
}

// OverlayUpdate implements ModalOverlay for UserFormDialogModel.
func (d *UserFormDialogModel) OverlayUpdate(msg tea.KeyMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for UserFormDialogModel.
func (d *UserFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// Render renders the user form dialog
func (d UserFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	// Username field
	content.WriteString(d.labelStyle.Render("Username"))
	content.WriteString("\n")
	if d.focusIndex == 0 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.UsernameInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.UsernameInput.View()))
	}
	content.WriteString("\n")

	// Name field
	content.WriteString(d.labelStyle.Render("Name"))
	content.WriteString("\n")
	if d.focusIndex == 1 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.NameInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.NameInput.View()))
	}
	content.WriteString("\n")

	// Email field
	content.WriteString(d.labelStyle.Render("Email"))
	content.WriteString("\n")
	if d.focusIndex == 2 {
		content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.EmailInput.View()))
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render(d.EmailInput.View()))
	}
	content.WriteString("\n")

	// Password field (create mode only)
	if d.isCreateMode() {
		content.WriteString(d.labelStyle.Render("Password"))
		content.WriteString("\n")
		if d.focusIndex == 3 {
			content.WriteString(d.focusedInputStyle.Width(innerW).Render(d.PasswordInput.View()))
		} else {
			content.WriteString(d.inputStyle.Width(innerW).Render(d.PasswordInput.View()))
		}
		content.WriteString("\n")
	}

	// Role carousel
	content.WriteString(d.labelStyle.Render("Role"))
	content.WriteString("\n")
	if len(d.RoleOptions) > 0 {
		optLabel := d.RoleOptions[d.RoleIndex].Label
		if d.focusIndex == d.roleFocusIndex() {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Primary).
				Background(config.DefaultStyle.Accent).
				Padding(0, 1).
				Render("◀ " + optLabel + " ▶")
			content.WriteString(selector)
		} else {
			selector := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Secondary).
				Padding(0, 1).
				Render("  " + optLabel + "  ")
			content.WriteString(selector)
		}
	} else {
		content.WriteString(d.inputStyle.Width(innerW).Render("No roles available"))
	}
	content.WriteString("\n\n")

	// Buttons
	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == d.cancelFocusIndex() {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == d.confirmFocusIndex() {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == d.cancelFocusIndex()))
	confirmBtn := confirmStyle.Render(buttonLabel("Save", d.focusIndex == d.confirmFocusIndex()))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// UserFormDialogAcceptMsg is sent when the user form dialog is confirmed.
type UserFormDialogAcceptMsg struct {
	Action   FormDialogAction
	EntityID string
	Username string
	Name     string
	Email    string
	Password string
	Role     string
}

// UserFormDialogCancelMsg is sent when the user form dialog is cancelled.
type UserFormDialogCancelMsg struct{}

// ShowUserFormDialogMsg triggers showing a user form dialog.
type ShowUserFormDialogMsg struct {
	Title string
	Roles []db.Roles
}

// ShowEditUserDialogMsg triggers showing an edit user dialog.
type ShowEditUserDialogMsg struct {
	User  db.UserWithRoleLabelRow
	Roles []db.Roles
}
