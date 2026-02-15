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
// The mutation and audit record are atomic -- both succeed or both fail.
//
// Hook integration (Phase 3): If AuditContext.HookRunner is non-nil and has
// hooks for before_create/after_create on this table, before-hooks run inside
// the transaction (can abort via error), after-hooks fire asynchronously after
// the transaction commits.
func Create[T any](cmd CreateCommand[T]) (T, error) {
	var result T
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	auditCtx := cmd.AuditContext()
	tableName := cmd.TableName()
	runner := auditCtx.HookRunner

	err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		created, err := cmd.Execute(ctx, tx)
		if err != nil {
			return fmt.Errorf("execute create: %w", err)
		}
		result = created

		// Phase 3: Run before_create hooks inside the transaction.
		// The hook can abort the entire operation by returning an error.
		if runner != nil && runner.HasHooks(HookBeforeCreate, tableName) {
			if hookErr := runner.RunBeforeHooks(ctx, HookBeforeCreate, tableName, created); hookErr != nil {
				return hookErr
			}
		}

		newValues, err := marshalToJSONData(created)
		if err != nil {
			return fmt.Errorf("marshal created entity: %w", err)
		}

		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    tableName,
			RecordID:     cmd.GetID(created),
			Operation:    types.OpInsert,
			Action:       types.ActionCreate,
			UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
			NewValues:    newValues,
			RequestID:    types.NullableString{String: auditCtx.RequestID, Valid: auditCtx.RequestID != ""},
			IP:           types.NullableString{String: auditCtx.IP, Valid: auditCtx.IP != ""},
		})
	})

	// Phase 3: Fire after_create hooks asynchronously after commit.
	// Only fires on success (err == nil). Uses the original context for metadata
	// but the hook engine uses its own shutdown context for execution.
	if err == nil && runner != nil && runner.HasHooks(HookAfterCreate, tableName) {
		runner.RunAfterHooks(ctx, HookAfterCreate, tableName, result)
	}

	return result, err
}

// Update executes an audited update operation.
// Captures before-state, executes update, records both states atomically.
//
// Hook integration (Phase 3): before_update hooks run inside the transaction
// with the before-state entity. DetectStatusTransition fires before_publish or
// before_archive hooks for content_data status changes (M12). After-hooks fire
// asynchronously after commit.
func Update[T any](cmd UpdateCommand[T]) error {
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	auditCtx := cmd.AuditContext()
	tableName := cmd.TableName()
	runner := auditCtx.HookRunner

	// Track state for after-hooks (need access outside the transaction closure).
	var beforeEntity T
	var extraBeforeEvents []HookEvent

	err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		before, err := cmd.GetBefore(ctx, tx)
		if err != nil {
			return fmt.Errorf("get before state: %w", err)
		}
		beforeEntity = before
		oldValues, err := marshalToJSONData(before)
		if err != nil {
			return fmt.Errorf("marshal before state: %w", err)
		}

		// Phase 3: Run before_update hooks inside the transaction.
		if runner != nil && runner.HasHooks(HookBeforeUpdate, tableName) {
			if hookErr := runner.RunBeforeHooks(ctx, HookBeforeUpdate, tableName, before); hookErr != nil {
				return hookErr
			}
		}

		// Phase 3 (M12): Detect status transitions for content_data.
		// Compare before-state with update params to detect publish/archive transitions.
		if runner != nil {
			beforeMap, mapErr := StructToMap(before)
			if mapErr == nil {
				paramsMap, paramsErr := StructToMap(cmd.Params())
				if paramsErr == nil {
					extraBeforeEvents = DetectStatusTransition(tableName, beforeMap, paramsMap)
					for _, extraEvent := range extraBeforeEvents {
						if runner.HasHooks(extraEvent, tableName) {
							if hookErr := runner.RunBeforeHooks(ctx, extraEvent, tableName, before); hookErr != nil {
								return hookErr
							}
						}
					}
				}
			}
		}

		if err := cmd.Execute(ctx, tx); err != nil {
			return fmt.Errorf("execute update: %w", err)
		}

		newValues, err := marshalToJSONData(cmd.Params())
		if err != nil {
			return fmt.Errorf("marshal update params: %w", err)
		}

		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    tableName,
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

	// Phase 3: Fire after-hooks asynchronously after commit.
	if err == nil && runner != nil {
		if runner.HasHooks(HookAfterUpdate, tableName) {
			runner.RunAfterHooks(ctx, HookAfterUpdate, tableName, beforeEntity)
		}
		// Fire after_publish/after_archive for detected status transitions.
		for _, beforeEvent := range extraBeforeEvents {
			afterEvent := BeforeToAfterEvent(beforeEvent)
			if runner.HasHooks(afterEvent, tableName) {
				runner.RunAfterHooks(ctx, afterEvent, tableName, beforeEntity)
			}
		}
	}

	return err
}

// Delete executes an audited delete operation.
// Captures before-state, executes delete, records deletion atomically.
//
// Hook integration (Phase 3): before_delete hooks run inside the transaction
// with the before-state entity. After-hooks fire asynchronously after commit.
func Delete[T any](cmd DeleteCommand[T]) error {
	ctx := cmd.Context()

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	auditCtx := cmd.AuditContext()
	tableName := cmd.TableName()
	runner := auditCtx.HookRunner

	// Track before-state for after-hooks.
	var beforeEntity T

	err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
		before, err := cmd.GetBefore(ctx, tx)
		if err != nil {
			return fmt.Errorf("get before state: %w", err)
		}
		beforeEntity = before
		oldValues, err := marshalToJSONData(before)
		if err != nil {
			return fmt.Errorf("marshal before state: %w", err)
		}

		// Phase 3: Run before_delete hooks inside the transaction.
		if runner != nil && runner.HasHooks(HookBeforeDelete, tableName) {
			if hookErr := runner.RunBeforeHooks(ctx, HookBeforeDelete, tableName, before); hookErr != nil {
				return hookErr
			}
		}

		if err := cmd.Execute(ctx, tx); err != nil {
			return fmt.Errorf("execute delete: %w", err)
		}

		return cmd.Recorder().Record(ctx, tx, ChangeEventParams{
			EventID:      types.NewEventID(),
			HlcTimestamp: types.HLCNow(),
			NodeID:       auditCtx.NodeID,
			TableName:    tableName,
			RecordID:     cmd.GetID(),
			Operation:    types.OpDelete,
			Action:       types.ActionDelete,
			UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
			OldValues:    oldValues,
			RequestID:    types.NullableString{String: auditCtx.RequestID, Valid: auditCtx.RequestID != ""},
			IP:           types.NullableString{String: auditCtx.IP, Valid: auditCtx.IP != ""},
		})
	})

	// Phase 3: Fire after_delete hooks asynchronously after commit.
	if err == nil && runner != nil && runner.HasHooks(HookAfterDelete, tableName) {
		runner.RunAfterHooks(ctx, HookAfterDelete, tableName, beforeEntity)
	}

	return err
}
