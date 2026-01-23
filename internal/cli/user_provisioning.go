package cli

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/crypto/bcrypt"
)

// UserProvisioningCompleteMsg is sent when user provisioning is complete
type UserProvisioningCompleteMsg struct {
	UserID types.UserID
	Error  error
}

// NewUserProvisioningForm creates a form for provisioning a new SSH user
func NewUserProvisioningForm(m Model) *huh.Form {
	var (
		username string
		name     string
		email    string
		password string
		confirm  string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Welcome to ModulaCMS!").
				Description(fmt.Sprintf(
					"Your SSH key is not yet registered.\n"+
						"Fingerprint: %s\n\n"+
						"Let's create your account:",
					m.SSHFingerprint,
				)),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Description("Your unique username (no spaces)").
				Value(&username).
				Validate(func(s string) error {
					if len(s) < 3 {
						return fmt.Errorf("username must be at least 3 characters")
					}
					return nil
				}),
			huh.NewInput().
				Title("Full Name").
				Description("Your full name").
				Value(&name).
				Validate(func(s string) error {
					if len(s) < 1 {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Email").
				Description("Your email address").
				Value(&email).
				Validate(func(s string) error {
					if len(s) < 3 || !utility.IsValidEmail(s) {
						return fmt.Errorf("please enter a valid email address")
					}
					return nil
				}),
		).Description("Account Details"),
		huh.NewGroup(
			huh.NewInput().
				Title("Password").
				Description("Choose a strong password").
				Value(&password).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if len(s) < 8 {
						return fmt.Errorf("password must be at least 8 characters")
					}
					return nil
				}),
			huh.NewInput().
				Title("Confirm Password").
				Description("Re-enter your password").
				Value(&confirm).
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s != password {
						return fmt.Errorf("passwords do not match")
					}
					return nil
				}),
		).Description("Security"),
	)

	form.Init()

	// Store the form values for later access
	m.FormState = &FormModel{
		Form:       form,
		FormValues: []*string{&username, &name, &email, &password},
	}

	return form
}

// ProvisionSSHUser creates a new user and registers their SSH key
func ProvisionSSHUser(m Model) tea.Cmd {
	return func() tea.Msg {
		dbc := db.ConfigDB(*m.Config)

		// Extract form values
		username := *m.FormState.FormValues[0]
		name := *m.FormState.FormValues[1]
		email := *m.FormState.FormValues[2]
		password := *m.FormState.FormValues[3]

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			utility.DefaultLogger.Error("Failed to hash password", err)
			return UserProvisioningCompleteMsg{Error: fmt.Errorf("failed to hash password: %v", err)}
		}

		// Check if user already exists
		existingUser, _ := dbc.GetUserByEmail(types.Email(email))
		if existingUser != nil {
			return UserProvisioningCompleteMsg{
				Error: fmt.Errorf("user with email %s already exists", email),
			}
		}

		// Create the user
		now := types.TimestampNow()
		user, err := dbc.CreateUser(db.CreateUserParams{
			Username:     username,
			Name:         name,
			Email:        types.Email(email),
			Hash:         string(hashedPassword),
			Role:         2, // Default role (adjust as needed)
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("Failed to create user", err)
			return UserProvisioningCompleteMsg{Error: fmt.Errorf("failed to create user: %v", err)}
		}

		utility.DefaultLogger.Info("Created user: %s (ID: %d)", user.Email, user.UserID)

		// Register the SSH key
		_, err = dbc.CreateUserSshKey(db.CreateUserSshKeyParams{
			UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
			PublicKey:   m.SSHPublicKey,
			KeyType:     m.SSHKeyType,
			Fingerprint: m.SSHFingerprint,
			Label:       fmt.Sprintf("Auto-provisioned key (%s)", time.Now().Format("2006-01-02")),
			DateCreated: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("Failed to register SSH key", err)
			return UserProvisioningCompleteMsg{
				UserID: user.UserID,
				Error:  fmt.Errorf("user created but failed to register SSH key: %v", err),
			}
		}

		utility.DefaultLogger.Info("Registered SSH key for user: %s (fingerprint: %s)", user.Email, m.SSHFingerprint)

		return UserProvisioningCompleteMsg{
			UserID: user.UserID,
			Error:  nil,
		}
	}
}
