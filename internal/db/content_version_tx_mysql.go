package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

func getMaxVersionNumberInTxMySQL(ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) (int64, error) {
	queries := mdbm.New(tx)
	result, err := queries.GetMaxVersionNumberForUpdate(ctx, mdbm.GetMaxVersionNumberForUpdateParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return 0, fmt.Errorf("tx get max version number: %w", err)
	}
	// MySQL sqlc generates int32 for INT columns; widen to int64
	return result.(int64), nil
}

func clearPublishedFlagInTxMySQL(ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) error {
	queries := mdbm.New(tx)
	return queries.ClearPublishedFlag(ctx, mdbm.ClearPublishedFlagParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
}

func createContentVersionInTxMySQL(d MysqlDatabase, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateContentVersionParams) (*ContentVersion, error) {
	cmd := MysqlDatabase{}.NewContentVersionCmd(ctx, ac, params)
	result, err := audited.CreateInTx(cmd, tx)
	if err != nil {
		return nil, fmt.Errorf("tx create content version: %w", err)
	}
	r := d.MapContentVersion(result)
	return &r, nil
}

func getAdminMaxVersionNumberInTxMySQL(ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) (int64, error) {
	queries := mdbm.New(tx)
	result, err := queries.GetAdminMaxVersionNumberForUpdate(ctx, mdbm.GetAdminMaxVersionNumberForUpdateParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return 0, fmt.Errorf("tx get admin max version number: %w", err)
	}
	return result.(int64), nil
}

func clearAdminPublishedFlagInTxMySQL(ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) error {
	queries := mdbm.New(tx)
	return queries.ClearAdminPublishedFlag(ctx, mdbm.ClearAdminPublishedFlagParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
}

func createAdminContentVersionInTxMySQL(d MysqlDatabase, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	cmd := MysqlDatabase{}.NewAdminContentVersionCmd(ctx, ac, params)
	result, err := audited.CreateInTx(cmd, tx)
	if err != nil {
		return nil, fmt.Errorf("tx create admin content version: %w", err)
	}
	r := d.MapAdminContentVersion(result)
	return &r, nil
}
