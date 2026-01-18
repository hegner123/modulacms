package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// UpdateProvisioning handles user provisioning flow
func (m Model) UpdateProvisioning(msg tea.Msg) (Model, tea.Cmd) {
	// If provisioning is complete, we're done
	if !m.NeedsProvisioning {
		return m, nil
	}

	// Initialize provisioning form if not already created
	if m.FormState == nil || m.FormState.Form == nil {
		form := NewUserProvisioningForm(m)
		m.FormState = &FormModel{Form: form}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case UserProvisioningCompleteMsg:
		if msg.Error != nil {
			// Show error message
			m.Status = ERROR
			m.Err = msg.Error
			m.NeedsProvisioning = false // Clear provisioning state
			return m, nil
		}

		// Success! Clear provisioning state and reload as authenticated user
		m.NeedsProvisioning = false
		m.FormState = nil

		// Show success message
		return m, ShowDialog(
			"Account Created!",
			"Your account has been created successfully.\nYou can now access ModulaCMS.",
			false,
		)
	}

	// Update the form
	form, cmd := m.FormState.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.FormState.Form = f

		if m.FormState.Form.State == huh.StateCompleted {
			// Form completed, provision the user
			return m, ProvisionSSHUser(m)
		}
	}


	return m, cmd
}
