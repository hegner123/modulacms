// Package ops provides generic sibling-pointer tree algorithms parameterized
// by ID type. It has no dependency on db, service, or config — only on
// internal/db/audited for the AuditContext type in Backend.UpdatePointers.
//
// ContentService implements Backend[types.ContentID] and AdminContentService
// implements Backend[types.AdminContentID]. A ContentID cannot be accidentally
// passed to an admin tree operation — the compiler rejects it.
package ops

import (
	"context"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// Node represents a node in a sibling-pointer tree. Generic over ID type
// to preserve compile-time safety between ContentID and AdminContentID.
type Node[ID ~string] struct {
	ID            ID
	ParentID      NullableID[ID]
	FirstChildID  NullableID[ID]
	NextSiblingID NullableID[ID]
	PrevSiblingID NullableID[ID]
}

// NullableID is a generic optional ID.
type NullableID[ID ~string] struct {
	Value ID
	Valid bool
}

// NullID returns a NullableID with Valid=true.
func NullID[ID ~string](id ID) NullableID[ID] {
	return NullableID[ID]{Value: id, Valid: true}
}

// EmptyID returns a NullableID with Valid=false.
func EmptyID[ID ~string]() NullableID[ID] {
	return NullableID[ID]{}
}

// Pointers holds only the pointer fields for an update.
type Pointers[ID ~string] struct {
	ParentID      NullableID[ID]
	FirstChildID  NullableID[ID]
	NextSiblingID NullableID[ID]
	PrevSiblingID NullableID[ID]
}

// Backend abstracts the database operations needed by tree algorithms.
// Implementations fetch and update sibling pointers on real content rows.
type Backend[ID ~string] interface {
	GetNode(ctx context.Context, id ID) (*Node[ID], error)
	UpdatePointers(ctx context.Context, ac audited.AuditContext, id ID, ptrs Pointers[ID]) error
}

// TxBackend extends Backend with transaction support.
// Algorithm functions accept Backend; callers can wrap in WithTx.
type TxBackend[ID ~string] interface {
	Backend[ID]
	WithTx(ctx context.Context, fn func(Backend[ID]) error) error
}

// MoveParams describes a tree move operation.
type MoveParams[ID ~string] struct {
	NodeID      ID
	NewParentID NullableID[ID]
	Position    int // 0 = first child, positive = insert at position
}

// MoveResult reports what happened during a move.
type MoveResult[ID ~string] struct {
	OldParentID NullableID[ID]
	NewParentID NullableID[ID]
	OldPosition int
	NewPosition int
	Report      *OperationReport[ID] // chain snapshots + validation + assertions
}

// ReorderResult reports what happened during a reorder.
type ReorderResult[ID ~string] struct {
	Updated int
	Report  *OperationReport[ID]
}

// SaveParams describes a bulk tree save (create/remap/update/delete).
type SaveParams[ID ~string] struct {
	ParentID NullableID[ID]
	Creates  []SaveCreate[ID]
	Updates  []SaveUpdate[ID]
	Deletes  []ID
	Order    []ID // final sibling order (includes temp IDs before remap)
}

// SaveCreate is a node to create during Save.
type SaveCreate[ID ~string] struct {
	TempID ID // client-assigned temporary ID
	// CreateFn is called by Save to create the node. Returns the real ID.
	CreateFn func(ctx context.Context, ac audited.AuditContext) (ID, error)
}

// SaveUpdate is a node to update during Save (non-pointer fields only).
type SaveUpdate[ID ~string] struct {
	ID       ID
	UpdateFn func(ctx context.Context, ac audited.AuditContext) error
}

// SaveResult reports what happened during a bulk save.
type SaveResult[ID ~string] struct {
	IDMap   map[ID]ID // temp ID -> real ID
	Created int
	Updated int
	Deleted int
	Report  *OperationReport[ID]
}
