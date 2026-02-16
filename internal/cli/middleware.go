package cli

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/muesli/termenv"
)

// timeMsg is a message type for periodic time updates.
type timeMsg time.Time

// CliMiddleware returns a Wish middleware that launches the CLI TUI application for SSH sessions.
func CliMiddleware(v *bool, c *config.Config, driver db.DbDriver, logger Logger, pluginMgr *plugin.Manager, mgr *config.Manager) wish.Middleware {
	newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
		p := tea.NewProgram(m, opts...)
		go func() {
			for {
				<-time.After(1 * time.Second)
				p.Send(timeMsg(time.Now()))
			}
		}()
		return p
	}
	teaHandler := func(s ssh.Session) *tea.Program {
		pty, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}
		m, _ := InitialModel(v, c, driver, logger, pluginMgr, mgr)
		m.Term = pty.Term
		m.Width = pty.Window.Width
		m.Height = pty.Window.Height
		m.Time = time.Now()

		// Always store SSH key info from the session
		if s.PublicKey() != nil {
			keyBytes := s.PublicKey().Marshal()
			hash := sha256.Sum256(keyBytes)
			m.SSHFingerprint = fmt.Sprintf("SHA256:%s", base64.StdEncoding.EncodeToString(hash[:]))
			m.SSHKeyType = s.PublicKey().Type()
			m.SSHPublicKey = base64.StdEncoding.EncodeToString(keyBytes)
		}

		// Check if user needs provisioning
		ctx := s.Context()
		if needsProvisioning, ok := ctx.Value("needs_provisioning").(bool); ok && needsProvisioning {
			m.NeedsProvisioning = true
		}

		// Extract authenticated user ID from SSH context
		if userID, ok := ctx.Value("user_id").(types.UserID); ok {
			m.UserID = userID
		}

		return newProg(&m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
