package ops

import (
	"context"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// UnlinkChecked wraps Unlink with before/after snapshots, pre-op validation,
// and post-op assertions. If pre-op validation finds mangled data, returns
// *ChainError and does not execute. If post-op assertions fail, returns
// *AssertionError (inside a transaction this triggers rollback).
func UnlinkChecked[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], node *Node[ID]) (*OperationReport[ID], error) {
	report := &OperationReport[ID]{}

	// Before snapshot: use parent's chain walk if the node has a parent,
	// otherwise snapshot the node and its direct neighbors.
	var before *ChainSnapshot[ID]
	if node.ParentID.Valid {
		var walkErr error
		before, walkErr = WalkSiblingChain(ctx, b, node.ParentID.Value)
		if walkErr != nil {
			return nil, walkErr
		}
	} else {
		affectedIDs := []ID{node.ID}
		if node.PrevSiblingID.Valid {
			affectedIDs = append(affectedIDs, node.PrevSiblingID.Value)
		}
		if node.NextSiblingID.Valid {
			affectedIDs = append(affectedIDs, node.NextSiblingID.Value)
		}
		var snapErr error
		before, snapErr = SnapshotNodes(ctx, b, affectedIDs)
		if snapErr != nil {
			return nil, snapErr
		}
	}
	report.Before = before

	// Pre-op validation
	violations := ValidateChain(before)
	if len(violations) > 0 {
		report.Violations = violations
		return report, &ChainError[ID]{Operation: "unlink", Report: report}
	}

	// Execute
	if err := Unlink(ctx, ac, b, node); err != nil {
		return report, err
	}

	// After snapshot: re-read all nodes that were in the before snapshot
	afterIDs := make([]ID, 0, len(before.Nodes))
	for _, n := range before.Nodes {
		afterIDs = append(afterIDs, n.ID)
	}
	after, err := SnapshotNodes(ctx, b, afterIDs)
	if err != nil {
		return report, err
	}
	report.After = after

	// Post-op assertions
	assertions := AssertChainConsistency(after)
	report.Assertions = assertions
	for _, a := range assertions {
		if !a.Passed {
			return report, &AssertionError[ID]{Operation: "unlink", Report: report}
		}
	}

	return report, nil
}

// InsertAtChecked wraps InsertAt with before/after snapshots, pre-op validation,
// and post-op assertions.
func InsertAtChecked[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], parentID, childID ID, position int) (*OperationReport[ID], error) {
	report := &OperationReport[ID]{}

	// Before snapshot: parent's chain only (child isn't attached yet)
	before, err := WalkSiblingChain(ctx, b, parentID)
	if err != nil {
		return nil, err
	}
	report.Before = before

	// Pre-op validation on the existing chain (without the child)
	violations := ValidateChain(before)
	if len(violations) > 0 {
		report.Violations = violations
		return report, &ChainError[ID]{Operation: "insert", Report: report}
	}

	// Execute
	if err := InsertAt(ctx, ac, b, parentID, childID, position); err != nil {
		return report, err
	}

	// After snapshot
	after, err := WalkSiblingChain(ctx, b, parentID)
	if err != nil {
		return report, err
	}
	report.After = after

	// Post-op assertions
	assertions := AssertChainConsistency(after)
	report.Assertions = assertions
	for _, a := range assertions {
		if !a.Passed {
			return report, &AssertionError[ID]{Operation: "insert", Report: report}
		}
	}

	return report, nil
}

// AppendChildChecked wraps AppendChild with before/after snapshots, pre-op
// validation, and post-op assertions.
func AppendChildChecked[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], parentID, childID ID) (*OperationReport[ID], error) {
	report := &OperationReport[ID]{}

	// Before snapshot: parent's chain only (child isn't attached yet)
	before, err := WalkSiblingChain(ctx, b, parentID)
	if err != nil {
		return nil, err
	}
	report.Before = before

	// Pre-op validation on the existing chain (without the child)
	violations := ValidateChain(before)
	if len(violations) > 0 {
		report.Violations = violations
		return report, &ChainError[ID]{Operation: "append", Report: report}
	}

	// Execute
	if err := AppendChild(ctx, ac, b, parentID, childID); err != nil {
		return report, err
	}

	// After snapshot
	after, err := WalkSiblingChain(ctx, b, parentID)
	if err != nil {
		return report, err
	}
	report.After = after

	// Post-op assertions
	assertions := AssertChainConsistency(after)
	report.Assertions = assertions
	for _, a := range assertions {
		if !a.Passed {
			return report, &AssertionError[ID]{Operation: "append", Report: report}
		}
	}

	return report, nil
}
