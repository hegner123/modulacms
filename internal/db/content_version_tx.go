package db

import (
	"context"
	"database/sql"
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// GetMaxVersionNumberInTx reads the highest version number within an existing transaction.
// MySQL/PostgreSQL variants use FOR UPDATE to prevent concurrent reads.
func GetMaxVersionNumberInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) (int64, error) {
	switch d.(type) {
	case Database:
		queries := mdb.New(tx)
		result, err := queries.GetMaxVersionNumberForUpdate(ctx, mdb.GetMaxVersionNumberForUpdateParams{
			ContentDataID: contentDataID,
			Locale:        locale,
		})
		if err != nil {
			return 0, fmt.Errorf("tx get max version number: %w", err)
		}
		return result.(int64), nil
	case MysqlDatabase:
		return getMaxVersionNumberInTxMySQL(ctx, tx, contentDataID, locale)
	case PsqlDatabase:
		return getMaxVersionNumberInTxPsql(ctx, tx, contentDataID, locale)
	default:
		return 0, fmt.Errorf("tx get max version number: unsupported driver type %T", d)
	}
}

// ClearPublishedFlagInTx clears published flags within an existing transaction.
func ClearPublishedFlagInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	contentDataID types.ContentID, locale string) error {
	switch d.(type) {
	case Database:
		queries := mdb.New(tx)
		return queries.ClearPublishedFlag(ctx, mdb.ClearPublishedFlagParams{
			ContentDataID: contentDataID,
			Locale:        locale,
		})
	case MysqlDatabase:
		return clearPublishedFlagInTxMySQL(ctx, tx, contentDataID, locale)
	case PsqlDatabase:
		return clearPublishedFlagInTxPsql(ctx, tx, contentDataID, locale)
	default:
		return fmt.Errorf("tx clear published flag: unsupported driver type %T", d)
	}
}

// CreateContentVersionInTx creates a content version with audit trail in an existing transaction.
func CreateContentVersionInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateContentVersionParams) (*ContentVersion, error) {
	switch drv := d.(type) {
	case Database:
		cmd := Database{}.NewContentVersionCmd(ctx, ac, params)
		result, err := audited.CreateInTx(cmd, tx)
		if err != nil {
			return nil, fmt.Errorf("tx create content version: %w", err)
		}
		r := drv.MapContentVersion(result)
		return &r, nil
	case MysqlDatabase:
		return createContentVersionInTxMySQL(drv, ctx, tx, ac, params)
	case PsqlDatabase:
		return createContentVersionInTxPsql(drv, ctx, tx, ac, params)
	default:
		return nil, fmt.Errorf("tx create content version: unsupported driver type %T", d)
	}
}

// GetAdminMaxVersionNumberInTx reads the highest admin version number within an existing transaction.
func GetAdminMaxVersionNumberInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) (int64, error) {
	switch d.(type) {
	case Database:
		queries := mdb.New(tx)
		result, err := queries.GetAdminMaxVersionNumberForUpdate(ctx, mdb.GetAdminMaxVersionNumberForUpdateParams{
			AdminContentDataID: adminContentDataID,
			Locale:             locale,
		})
		if err != nil {
			return 0, fmt.Errorf("tx get admin max version number: %w", err)
		}
		return result.(int64), nil
	case MysqlDatabase:
		return getAdminMaxVersionNumberInTxMySQL(ctx, tx, adminContentDataID, locale)
	case PsqlDatabase:
		return getAdminMaxVersionNumberInTxPsql(ctx, tx, adminContentDataID, locale)
	default:
		return 0, fmt.Errorf("tx get admin max version number: unsupported driver type %T", d)
	}
}

// ClearAdminPublishedFlagInTx clears admin published flags within an existing transaction.
func ClearAdminPublishedFlagInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	adminContentDataID types.AdminContentID, locale string) error {
	switch d.(type) {
	case Database:
		queries := mdb.New(tx)
		return queries.ClearAdminPublishedFlag(ctx, mdb.ClearAdminPublishedFlagParams{
			AdminContentDataID: adminContentDataID,
			Locale:             locale,
		})
	case MysqlDatabase:
		return clearAdminPublishedFlagInTxMySQL(ctx, tx, adminContentDataID, locale)
	case PsqlDatabase:
		return clearAdminPublishedFlagInTxPsql(ctx, tx, adminContentDataID, locale)
	default:
		return fmt.Errorf("tx clear admin published flag: unsupported driver type %T", d)
	}
}

// CreateAdminContentVersionInTx creates an admin content version with audit trail in an existing transaction.
func CreateAdminContentVersionInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
	ac audited.AuditContext, params CreateAdminContentVersionParams) (*AdminContentVersion, error) {
	switch drv := d.(type) {
	case Database:
		cmd := Database{}.NewAdminContentVersionCmd(ctx, ac, params)
		result, err := audited.CreateInTx(cmd, tx)
		if err != nil {
			return nil, fmt.Errorf("tx create admin content version: %w", err)
		}
		r := drv.MapAdminContentVersion(result)
		return &r, nil
	case MysqlDatabase:
		return createAdminContentVersionInTxMySQL(drv, ctx, tx, ac, params)
	case PsqlDatabase:
		return createAdminContentVersionInTxPsql(drv, ctx, tx, ac, params)
	default:
		return nil, fmt.Errorf("tx create admin content version: unsupported driver type %T", d)
	}
}
