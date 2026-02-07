package audited

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
)

// marshalToJSONData serializes v to JSON and wraps it in a types.JSONData.
func marshalToJSONData(v any) (types.JSONData, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return types.JSONData{}, err
	}
	return types.NewJSONData(json.RawMessage(b)), nil
}

// Create executes an audited create operation.
// The mutation and audit record are atomic — both succeed or both fail.
func Create[T any](cmd CreateCommand[T]) (T, error) {
	var result T
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		created, err := cmd.Execute(ctx, tx)
		if err != nil {
			return fmt.Errorf("execute create: %w", err)
		}
		result = created

		newValues, err := marshalToJSONData(created)
		if err != nil {
			return fmt.Errorf("marshal created entity: %w", err)
		}

		auditCtx := cmd.AuditContext()
		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    cmd.TableName(),
			RecordID:     cmd.GetID(created),
			Operation:    types.OpInsert,
			Action:       types.ActionCreate,
			UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
			NewValues:    newValues,
			RequestID:    types.NullableString{String: auditCtx.RequestID, Valid: auditCtx.RequestID != ""},
			IP:           types.NullableString{String: auditCtx.IP, Valid: auditCtx.IP != ""},
		})
	})

	return result, err
}

// Update executes an audited update operation.
// Captures before-state, executes update, records both states atomically.
func Update[T any](cmd UpdateCommand[T]) error {
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	return types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		before, err := cmd.GetBefore(ctx, tx)
		if err != nil {
			return fmt.Errorf("get before state: %w", err)
		}
		oldValues, err := marshalToJSONData(before)
		if err != nil {
			return fmt.Errorf("marshal before state: %w", err)
		}

		if err := cmd.Execute(ctx, tx); err != nil {
			return fmt.Errorf("execute update: %w", err)
		}

		newValues, err := marshalToJSONData(cmd.Params())
		if err != nil {
			return fmt.Errorf("marshal update params: %w", err)
		}

		auditCtx := cmd.AuditContext()
		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    cmd.TableName(),
			RecordID:     cmd.GetID(),
			Operation:    types.OpUpdate,
			Action:       types.ActionUpdate,
			UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
			OldValues:    oldValues,
			NewValues:    newValues,
			RequestID:    types.NullableString{String: auditCtx.RequestID, Valid: auditCtx.RequestID != ""},
			IP:           types.NullableString{String: auditCtx.IP, Valid: auditCtx.IP != ""},
		})
	})
}

// Delete executes an audited delete operation.
// Captures before-state, executes delete, records deletion atomically.
func Delete[T any](cmd DeleteCommand[T]) error {
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	return types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		before, err := cmd.GetBefore(ctx, tx)
		if err != nil {
			return fmt.Errorf("get before state: %w", err)
		}
		oldValues, err := marshalToJSONData(before)
		if err != nil {
			return fmt.Errorf("marshal before state: %w", err)
		}

		if err := cmd.Execute(ctx, tx); err != nil {
			return fmt.Errorf("execute delete: %w", err)
		}

		auditCtx := cmd.AuditContext()
		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    cmd.TableName(),
			RecordID:     cmd.GetID(),
			Operation:    types.OpDelete,
			Action:       types.ActionDelete,
			UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
			OldValues:    oldValues,
			RequestID:    types.NullableString{String: auditCtx.RequestID, Valid: auditCtx.RequestID != ""},
			IP:           types.NullableString{String: auditCtx.IP, Valid: auditCtx.IP != ""},
		})
	})
}
