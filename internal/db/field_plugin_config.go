package db

import (
	"context"
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	"github.com/hegner123/modulacms/internal/db/types"
)

// FieldPluginConfig is the application-level type for the field_plugin_config
// extension table. Binds a field of type "plugin" to a specific plugin and
// interface definition.
type FieldPluginConfig struct {
	FieldID         types.FieldID
	PluginName      string
	PluginInterface string
	PluginVersion   string
	DateCreated     types.Timestamp
	DateModified    types.Timestamp
}

// CreateFieldPluginConfigParams holds parameters for creating a field plugin config.
type CreateFieldPluginConfigParams struct {
	FieldID         types.FieldID
	PluginName      string
	PluginInterface string
	PluginVersion   string
	DateCreated     types.Timestamp
	DateModified    types.Timestamp
}

// UpdateFieldPluginConfigParams holds parameters for updating a field plugin config.
type UpdateFieldPluginConfigParams struct {
	FieldID         types.FieldID
	PluginName      string
	PluginInterface string
	PluginVersion   string
	DateModified    types.Timestamp
}

// --- SQLite (Database) wrapper methods ---

func (d Database) CreateFieldPluginConfigTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateFieldPluginConfigTable(d.Context)
}

func (d Database) CreateAdminFieldPluginConfigTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateAdminFieldPluginConfigTable(d.Context)
}

func (d Database) GetFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetFieldPluginConfig(ctx, mdb.GetFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get field plugin config: %w", err)
	}
	return mapFieldPluginConfig(row), nil
}

func (d Database) GetAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminFieldPluginConfig(ctx, mdb.GetAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get admin field plugin config: %w", err)
	}
	return mapAdminFieldPluginConfig(row), nil
}

func (d Database) CreateFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdb.New(d.Connection)
	return queries.CreateFieldPluginConfig(ctx, mdb.CreateFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d Database) CreateAdminFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdb.New(d.Connection)
	return queries.CreateAdminFieldPluginConfig(ctx, mdb.CreateAdminFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d Database) UpdateFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateFieldPluginConfig(ctx, mdb.UpdateFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d Database) UpdateAdminFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateAdminFieldPluginConfig(ctx, mdb.UpdateAdminFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d Database) DeleteFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteFieldPluginConfig(ctx, mdb.DeleteFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

func (d Database) DeleteAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteAdminFieldPluginConfig(ctx, mdb.DeleteAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

// --- MySQL (MysqlDatabase) wrapper methods ---

func (d MysqlDatabase) CreateFieldPluginConfigTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateFieldPluginConfigTable(d.Context)
}

func (d MysqlDatabase) CreateAdminFieldPluginConfigTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateAdminFieldPluginConfigTable(d.Context)
}

func (d MysqlDatabase) GetFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetFieldPluginConfig(ctx, mdbm.GetFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get field plugin config: %w", err)
	}
	return mapMysqlFieldPluginConfig(row), nil
}

func (d MysqlDatabase) GetAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminFieldPluginConfig(ctx, mdbm.GetAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get admin field plugin config: %w", err)
	}
	return mapMysqlAdminFieldPluginConfig(row), nil
}

func (d MysqlDatabase) CreateFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdbm.New(d.Connection)
	return queries.CreateFieldPluginConfig(ctx, mdbm.CreateFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d MysqlDatabase) CreateAdminFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdbm.New(d.Connection)
	return queries.CreateAdminFieldPluginConfig(ctx, mdbm.CreateAdminFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d MysqlDatabase) UpdateFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateFieldPluginConfig(ctx, mdbm.UpdateFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d MysqlDatabase) UpdateAdminFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateAdminFieldPluginConfig(ctx, mdbm.UpdateAdminFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d MysqlDatabase) DeleteFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteFieldPluginConfig(ctx, mdbm.DeleteFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

func (d MysqlDatabase) DeleteAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteAdminFieldPluginConfig(ctx, mdbm.DeleteAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

// --- PostgreSQL (PsqlDatabase) wrapper methods ---

func (d PsqlDatabase) CreateFieldPluginConfigTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateFieldPluginConfigTable(d.Context)
}

func (d PsqlDatabase) CreateAdminFieldPluginConfigTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateAdminFieldPluginConfigTable(d.Context)
}

func (d PsqlDatabase) GetFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetFieldPluginConfig(ctx, mdbp.GetFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get field plugin config: %w", err)
	}
	return mapPsqlFieldPluginConfig(row), nil
}

func (d PsqlDatabase) GetAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) (*FieldPluginConfig, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminFieldPluginConfig(ctx, mdbp.GetAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
	if err != nil {
		return nil, fmt.Errorf("get admin field plugin config: %w", err)
	}
	return mapPsqlAdminFieldPluginConfig(row), nil
}

func (d PsqlDatabase) CreateFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdbp.New(d.Connection)
	return queries.CreateFieldPluginConfig(ctx, mdbp.CreateFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d PsqlDatabase) CreateAdminFieldPluginConfig(ctx context.Context, params CreateFieldPluginConfigParams) error {
	queries := mdbp.New(d.Connection)
	return queries.CreateAdminFieldPluginConfig(ctx, mdbp.CreateAdminFieldPluginConfigParams{
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateCreated:     params.DateCreated,
		DateModified:    params.DateModified,
	})
}

func (d PsqlDatabase) UpdateFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateFieldPluginConfig(ctx, mdbp.UpdateFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d PsqlDatabase) UpdateAdminFieldPluginConfig(ctx context.Context, params UpdateFieldPluginConfigParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateAdminFieldPluginConfig(ctx, mdbp.UpdateAdminFieldPluginConfigParams{
		PluginName:      params.PluginName,
		PluginInterface: params.PluginInterface,
		PluginVersion:   params.PluginVersion,
		DateModified:    params.DateModified,
		FieldID:         types.NullableFieldID{ID: params.FieldID, Valid: true},
	})
}

func (d PsqlDatabase) DeleteFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteFieldPluginConfig(ctx, mdbp.DeleteFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

func (d PsqlDatabase) DeleteAdminFieldPluginConfig(ctx context.Context, fieldID types.FieldID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteAdminFieldPluginConfig(ctx, mdbp.DeleteAdminFieldPluginConfigParams{FieldID: types.NullableFieldID{ID: fieldID, Valid: true}})
}

// --- Mappers (SQLite) ---

func mapFieldPluginConfig(row mdb.FieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     types.Timestamp(row.DateCreated),
		DateModified:    types.Timestamp(row.DateModified),
	}
}

func mapAdminFieldPluginConfig(row mdb.AdminFieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     types.Timestamp(row.DateCreated),
		DateModified:    types.Timestamp(row.DateModified),
	}
}

// --- Mappers (MySQL) ---

func mapMysqlFieldPluginConfig(row mdbm.FieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     row.DateCreated,
		DateModified:    row.DateModified,
	}
}

func mapMysqlAdminFieldPluginConfig(row mdbm.AdminFieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     row.DateCreated,
		DateModified:    row.DateModified,
	}
}

// --- Mappers (PostgreSQL) ---

func mapPsqlFieldPluginConfig(row mdbp.FieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     row.DateCreated,
		DateModified:    row.DateModified,
	}
}

func mapPsqlAdminFieldPluginConfig(row mdbp.AdminFieldPluginConfig) *FieldPluginConfig {
	return &FieldPluginConfig{
		FieldID:         row.FieldID.ID,
		PluginName:      row.PluginName,
		PluginInterface: row.PluginInterface,
		PluginVersion:   row.PluginVersion,
		DateCreated:     row.DateCreated,
		DateModified:    row.DateModified,
	}
}
