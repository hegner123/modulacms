package install

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Step represents a single installation step
type Step struct {
	Name        string
	Description string
	Action      func() error
}

// InstallProgress manages a sequence of installation steps with progress indicators
type InstallProgress struct {
	steps        []Step
	successStyle lipgloss.Style
	failStyle    lipgloss.Style
	stepStyle    lipgloss.Style
}

// NewInstallProgress creates a new progress tracker
func NewInstallProgress() *InstallProgress {
	return &InstallProgress{
		steps:        []Step{},
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // green
		failStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // red
		stepStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")), // gray
	}
}

// AddStep adds a step to the progress tracker
func (p *InstallProgress) AddStep(name, description string, action func() error) *InstallProgress {
	p.steps = append(p.steps, Step{
		Name:        name,
		Description: description,
		Action:      action,
	})
	return p
}

// Run executes all steps with spinner feedback
func (p *InstallProgress) Run() error {
	total := len(p.steps)

	for i, step := range p.steps {
		stepNum := i + 1
		prefix := fmt.Sprintf("[%d/%d]", stepNum, total)

		// Create spinner for this step
		var stepErr error
		action := func(ctx context.Context) error {
			stepErr = step.Action()
			return stepErr
		}

		s := spinner.New().
			Title(fmt.Sprintf(" %s %s...", prefix, step.Description)).
			Type(spinner.Dots).
			ActionWithErr(action)

		err := s.Run()
		if err != nil {
			// Spinner itself failed (rare)
			fmt.Printf("  %s %s\n", p.failStyle.Render("X"), step.Name)
			return err
		}

		if stepErr != nil {
			// Step action failed
			fmt.Printf("  %s %s\n", p.failStyle.Render("X"), step.Name)
			return stepErr
		}

		// Success
		fmt.Printf("  %s %s\n", p.successStyle.Render("\u2713"), step.Name)
	}

	return nil
}

// RunWithWarnings executes all steps, collecting warnings but not stopping on non-critical failures
func (p *InstallProgress) RunWithWarnings() (warnings []string, err error) {
	total := len(p.steps)

	for i, step := range p.steps {
		stepNum := i + 1
		prefix := fmt.Sprintf("[%d/%d]", stepNum, total)

		var stepErr error
		action := func(ctx context.Context) error {
			stepErr = step.Action()
			return stepErr
		}

		s := spinner.New().
			Title(fmt.Sprintf(" %s %s...", prefix, step.Description)).
			Type(spinner.Dots).
			ActionWithErr(action)

		spinErr := s.Run()
		if spinErr != nil {
			fmt.Printf("  %s %s\n", p.failStyle.Render("X"), step.Name)
			return warnings, spinErr
		}

		if stepErr != nil {
			// Collect as warning, continue
			fmt.Printf("  %s %s (warning: %v)\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("!"),
				step.Name,
				stepErr)
			warnings = append(warnings, fmt.Sprintf("%s: %v", step.Name, stepErr))
			continue
		}

		fmt.Printf("  %s %s\n", p.successStyle.Render("\u2713"), step.Name)
	}

	return warnings, nil
}

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	fmt.Printf("\n%s %s\n", style.Render("\u2713"), msg)
}

// PrintError prints an error message
func PrintError(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	fmt.Printf("\n%s %s\n", style.Render("X"), msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	fmt.Printf("\n%s %s\n", style.Render("!"), msg)
}
