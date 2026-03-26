package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

func getContentDataInTxMySQL(d MysqlDatabase, ctx context.Context, tx *sql.Tx, id types.ContentID) (*ContentData, error) {
	queries := mdbm.New(tx)
	row, err := queries.GetContentData(ctx, mdbm.GetContentDataParams{ContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("tx get content_data %s: %w", id, err)
	}
	res := d.MapContentData(row)
	return &res, nil
}

func updateContentDataInTxMySQL(ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateContentDataParams) error {
	cmd := MysqlDatabase{}.UpdateContentDataCmd(ctx, ac, params)
	return audited.UpdateInTx(cmd, tx)
}

func getAdminContentDataInTxMySQL(d MysqlDatabase, ctx context.Context, tx *sql.Tx, id types.AdminContentID) (*AdminContentData, error) {
	queries := mdbm.New(tx)
	row, err := queries.GetAdminContentData(ctx, mdbm.GetAdminContentDataParams{AdminContentDataID: id})
	if err != nil {
		return nil, fmt.Errorf("tx get admin_content_data %s: %w", id, err)
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}

func updateAdminContentDataInTxMySQL(ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateAdminContentDataParams) error {
	cmd := MysqlDatabase{}.UpdateAdminContentDataCmd(ctx, ac, params)
	return audited.UpdateInTx(cmd, tx)
}
