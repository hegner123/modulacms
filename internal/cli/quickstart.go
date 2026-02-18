package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/definitions"
	"github.com/hegner123/modulacms/internal/middleware"
)

// QuickstartConfirmMsg is sent when the user selects a schema definition to install.
type QuickstartConfirmMsg struct {
	SchemaIndex int
}

// QuickstartMenuLabels builds a label list from the definitions registry.
func QuickstartMenuLabels() []string {
	defs := definitions.List()
	labels := make([]string, len(defs))
	for i, def := range defs {
		labels[i] = fmt.Sprintf("%s (%s)", def.Label, def.Name)
	}
	return labels
}

// dbInstallerAdapter bridges the definitions.Installer interface (no context, no error)
// with the db.DbDriver interface (context + audit context + error return).
type dbInstallerAdapter struct {
	driver db.DbDriver
	ctx    context.Context
	ac     audited.AuditContext
}

func (a *dbInstallerAdapter) CreateDatatype(p db.CreateDatatypeParams) db.Datatypes {
	result, err := a.driver.CreateDatatype(a.ctx, a.ac, p)
	if err != nil || result == nil {
		return db.Datatypes{}
	}
	return *result
}

func (a *dbInstallerAdapter) CreateField(p db.CreateFieldParams) db.Fields {
	result, err := a.driver.CreateField(a.ctx, a.ac, p)
	if err != nil || result == nil {
		return db.Fields{}
	}
	return *result
}

func (a *dbInstallerAdapter) CreateDatatypeField(p db.CreateDatatypeFieldParams) db.DatatypeFields {
	result, err := a.driver.CreateDatatypeField(a.ctx, a.ac, p)
	if err != nil || result == nil {
		return db.DatatypeFields{}
	}
	return *result
}

// RunQuickstartInstallCmd creates a tea.Cmd that installs a schema definition by index.
func RunQuickstartInstallCmd(cfg *config.Config, userID types.UserID, schemaIndex int) tea.Cmd {
	return func() tea.Msg {
		defs := definitions.List()
		if schemaIndex < 0 || schemaIndex >= len(defs) {
			return ActionResultMsg{
				Title:   "Install Failed",
				Message: "Invalid schema selection.",
				IsError: true,
			}
		}

		def := defs[schemaIndex]

		driver := db.ConfigDB(*cfg)
		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		adapter := &dbInstallerAdapter{
			driver: driver,
			ctx:    ctx,
			ac:     ac,
		}

		result, err := definitions.Install(adapter, def, userID)
		if err != nil {
			return ActionResultMsg{
				Title:   "Install Failed",
				Message: fmt.Sprintf("Failed to install %q:\n%s", def.Label, err),
				IsError: true,
			}
		}

		return ActionResultMsg{
			Title: "Install Complete",
			Message: fmt.Sprintf("Installed %q successfully.\n\nDatatypes: %d\nFields: %d\nJunction links: %d",
				result.DefinitionName, result.Datatypes, result.Fields, result.JunctionLinks),
		}
	}
}
