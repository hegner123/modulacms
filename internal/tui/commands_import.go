package tui

import (
	"context"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// RunImportCmd reads the file and executes the import via the service layer.
func RunImportCmd(ctx AppContext, format config.OutputFormat, path string) tea.Cmd {
	cfg := ctx.Config
	userID := ctx.UserID
	configMgr := ctx.ConfigManager

	return func() tea.Msg {
		body, err := os.ReadFile(path)
		if err != nil {
			return ImportCompleteMsg{Err: fmt.Errorf("read file: %w", err)}
		}

		d := db.ConfigDB(*cfg)
		bgCtx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		importSvc := service.NewImportService(d, configMgr)

		input := service.ImportContentInput{
			Format:  format,
			Body:    body,
			RouteID: types.NullableRouteID{},
		}

		result, importErr := importSvc.ImportContent(bgCtx, ac, input)
		if importErr != nil {
			return ImportCompleteMsg{Err: fmt.Errorf("import failed: %w", importErr)}
		}

		return ImportCompleteMsg{Result: result}
	}
}
