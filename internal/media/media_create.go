package media

import (
	"context"
	"fmt"

	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// CreateMedia creates a new media record with the given name using the provided config.
func CreateMedia(name string, c config.Config) (string, error) {
	d := db.ConfigDB(c)
	params := db.CreateMediaParams{
		Name: db.NewNullString(name),
	}
	ctx := context.Background()
	ac := audited.Ctx(types.NodeID(c.Node_ID), types.UserID(""), "", "system")
	mediaRow, err := d.CreateMedia(ctx, ac, params)
	if err != nil {
		return "", fmt.Errorf("failed to create media: %w", err)
	}
	return mediaRow.Name.String, nil
}
