package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/muesli/termenv"
)

type timeMsg time.Time

// TuiMiddleware returns a wish.Middleware that launches the new TUI
// for SSH sessions.
func TuiMiddleware(v *bool, c *config.Config) wish.Middleware {
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

		m, _ := InitialModel(v, c)
		m.Term = pty.Term
		m.Width = pty.Window.Width
		m.Height = pty.Window.Height
		m.Time = time.Now()

		return newProg(&m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
	}

	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
