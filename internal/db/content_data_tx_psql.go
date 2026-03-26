package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

func getContentDataInTxPsql(d PsqlDatabase, ctx context.Context, tx *sql.Tx, id types.ContentID) (*ContentData, error) {
	queries := mdbp.New(tx)
	row, err := queries.GetContentData(ctx, mdbp.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("tx get content_data %s: %w", id, err)
	}
	res := d.MapContentData(row)
	return &res, nil
}

func updateContentDataInTxPsql(ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateContentDataParams) error {
	cmd := PsqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)
	return audited.UpdateInTx(cmd, tx)
}

func getAdminContentDataInTxPsql(d PsqlDatabase, ctx context.Context, tx *sql.Tx, id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbp.New(tx)
	row, err := queries.GetAdminContentData(ctx, mdbp.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("tx get admin_content_data %s: %w", id, err)
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}

func updateAdminContentDataInTxPsql(ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateAdminContentDataParams) error {
	cmd := PsqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)
	return audited.UpdateInTx(cmd, tx)
}
