package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

func getMaxVersionNumberInTxPsql(ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) (int64, error) {
	queries := mdbp.New(tx)
	result, err := queries.GetMaxVersionNumberForUpdate(ctx, mdbp.GetMaxVersionNumberForUpdateParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
	if err != nil {
		return 0, fmt.Errorf("tx get max version number: %w", err)
	}
	return result.(int64), nil
}

func clearPublishedFlagInTxPsql(ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) error {
	queries := mdbp.New(tx)
	return queries.ClearPublishedFlag(ctx, mdbp.ClearPublishedFlagParams{
		ContentDataID: contentDataID,
		Locale:        locale,
	})
}

func createContentVersionInTxPsql(d PsqlDatabase, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateContentVersionParams) (*ContentVersion, error) {
	cmd := PsqlDatabase{}.NewContentVersionCmd(ctx, ac, params)
	result, err := audited.CreateInTx(cmd, tx)
	if err != nil {
		return nil, fmt.Errorf("tx create content version: %w", err)
	}
	r := d.MapContentVersion(result)
	return &r, nil
}

func getAdminMaxVersionNumberInTxPsql(ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) (int64, error) {
	queries := mdbp.New(tx)
	result, err := queries.GetAdminMaxVersionNumberForUpdate(ctx, mdbp.GetAdminMaxVersionNumberForUpdateParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
	if err != nil {
		return 0, fmt.Errorf("tx get admin max version number: %w", err)
	}
	return result.(int64), nil
}

func clearAdminPublishedFlagInTxPsql(ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) error {
	queries := mdbp.New(tx)
	return queries.ClearAdminPublishedFlag(ctx, mdbp.ClearAdminPublishedFlagParams{
		AdminContentDataID: adminContentDataID,
		Locale:             locale,
	})
}

func createAdminContentVersionInTxPsql(d PsqlDatabase, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	cmd := PsqlDatabase{}.NewAdminContentVersionCmd(ctx, ac, params)
	result, err := audited.CreateInTx(cmd, tx)
	if err != nil {
		return nil, fmt.Errorf("tx create admin content version: %w", err)
	}
	r := d.MapAdminContentVersion(result)
	return &r, nil
}
