package db

import (
	"context"
	"database/sql"
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// GetContentDataInTx reads a content data row within the provided transaction.
// Uses type-switching on the driver to select the correct sqlc package.
func GetContentDataInTx(d DbDriver, ctx context.Context, tx *sql.Tx, id types.ContentID) (*ContentData, error) {
	switch drv := d.(type) {
	case Database:
		queries := mdb.New(tx)
		row, err := queries.GetContentData(ctx, mdb.GetContentDataParams{ContentDataID: id})
		if err != nil {
			return nil, fmt.Errorf("tx get content_data %s: %w", id, err)
		}
		res := drv.MapContentData(row)
		return &res, nil
	case MysqlDatabase:
		return getContentDataInTxMySQL(drv, ctx, tx, id)
	case PsqlDatabase:
		return getContentDataInTxPsql(drv, ctx, tx, id)
	default:
		return nil, fmt.Errorf("tx get content_data: unsupported driver type %T", d)
	}
}

// UpdateContentDataInTx updates a content data row within the provided transaction,
// including audited change event recording. The update and its audit record are
// part of the caller's transaction — they commit or rollback together.
func UpdateContentDataInTx(d DbDriver, ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateContentDataParams) error {
	switch d.(type) {
	case Database:
		cmd := Database{}.UpdateContentDataCmd(ctx, ac, params)
		return audited.UpdateInTx(cmd, tx)
	case MysqlDatabase:
		return updateContentDataInTxMySQL(ctx, tx, ac, params)
	case PsqlDatabase:
		return updateContentDataInTxPsql(ctx, tx, ac, params)
	default:
		return fmt.Errorf("tx update content_data: unsupported driver type %T", d)
	}
}

// GetAdminContentDataInTx reads an admin content data row within the provided transaction.
func GetAdminContentDataInTx(d DbDriver, ctx context.Context, tx *sql.Tx, id types.AdminContentID) (*AdminContentData, error) {
	switch drv := d.(type) {
	case Database:
		queries := mdb.New(tx)
		row, err := queries.GetAdminContentData(ctx, mdb.GetAdminContentDataParams{AdminContentDataID: id})
		if err != nil {
			return nil, fmt.Errorf("tx get admin_content_data %s: %w", id, err)
		}
		res := drv.MapAdminContentData(row)
		return &res, nil
	case MysqlDatabase:
		return getAdminContentDataInTxMySQL(drv, ctx, tx, id)
	case PsqlDatabase:
		return getAdminContentDataInTxPsql(drv, ctx, tx, id)
	default:
		return nil, fmt.Errorf("tx get admin_content_data: unsupported driver type %T", d)
	}
}

// UpdateAdminContentDataInTx updates an admin content data row within the provided transaction.
func UpdateAdminContentDataInTx(d DbDriver, ctx context.Context, tx *sql.Tx, ac audited.AuditContext, params UpdateAdminContentDataParams) error {
	switch d.(type) {
	case Database:
		cmd := Database{}.UpdateAdminContentDataCmd(ctx, ac, params)
		return audited.UpdateInTx(cmd, tx)
	case MysqlDatabase:
		return updateAdminContentDataInTxMySQL(ctx, tx, ac, params)
	case PsqlDatabase:
		return updateAdminContentDataInTxPsql(ctx, tx, ac, params)
	default:
		return fmt.Errorf("tx update admin_content_data: unsupported driver type %T", d)
	}
}
