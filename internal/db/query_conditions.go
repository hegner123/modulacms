package db

import (
	"fmt"
	"strings"
)

// Safety limits for condition trees. These prevent resource exhaustion from
// deeply nested or excessively wide condition trees built by plugins.
const (
	MaxConditionDepth = 10  // max nesting of And/Or/Not
	MaxConditionNodes = 50  // total Condition nodes in a single tree
	MaxInValues       = 500 // max elements in a single IN clause
)

// Condition represents a WHERE clause expression that can be built into SQL.
// Security invariant: every Build implementation MUST:
// 1. Validate all column names via ValidColumnName
// 2. Quote all identifiers via quoteIdent
// 3. Use placeholder() for all user-provided values
// 4. Never embed raw string values into the SQL output
//
// Build accepts a BuildContext that tracks depth and total node count
// to prevent resource exhaustion from deeply nested or excessively wide
// condition trees.
type Condition interface {
	Build(ctx BuildContext, d Dialect, argOffset int) (sql string, args []any, nextOffset int, err error)
}

// BuildContext tracks depth and node count during condition tree traversal.
// NOT goroutine-safe. Each top-level query function creates its own BuildContext.
type BuildContext struct {
	CurrentDepth int
	NodeCount    *int // pointer: shared across recursive calls within one tree
}

// NewBuildContext creates a fresh BuildContext for a new condition tree traversal.
func NewBuildContext() BuildContext {
	count := 0
	return BuildContext{
		CurrentDepth: 0,
		NodeCount:    &count,
	}
}

// incrementNode increments the node count and returns an error if the limit is exceeded.
func (bc BuildContext) incrementNode() error {
	*bc.NodeCount++
	if *bc.NodeCount > MaxConditionNodes {
		return fmt.Errorf("condition tree exceeds maximum node count (%d)", MaxConditionNodes)
	}
	return nil
}

// checkDepth returns an error if the current depth exceeds the limit.
func (bc BuildContext) checkDepth() error {
	if bc.CurrentDepth > MaxConditionDepth {
		return fmt.Errorf("condition tree exceeds maximum depth (%d)", MaxConditionDepth)
	}
	return nil
}

// CompareOp is an allowlisted comparison operator.
type CompareOp string

const (
	OpEq   CompareOp = "="
	OpNeq  CompareOp = "<>"
	OpGt   CompareOp = ">"
	OpLt   CompareOp = "<"
	OpGte  CompareOp = ">="
	OpLte  CompareOp = "<="
	OpLike CompareOp = "LIKE"
)

var validCompareOps = map[CompareOp]bool{
	OpEq: true, OpNeq: true, OpGt: true, OpLt: true,
	OpGte: true, OpLte: true, OpLike: true,
}

// Compare represents a column comparison: "col" op ?.
type Compare struct {
	Column string
	Op     CompareOp
	Value  any
}

func (c Compare) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if !validCompareOps[c.Op] {
		return "", nil, 0, fmt.Errorf("invalid compare operator %q", c.Op)
	}
	if c.Value == nil {
		return "", nil, 0, fmt.Errorf("compare with %s requires a non-nil value; use IsNullCondition or IsNotNullCondition instead", c.Op)
	}
	if err := ValidColumnName(c.Column); err != nil {
		return "", nil, 0, fmt.Errorf("invalid column %q: %w", c.Column, err)
	}
	sql := fmt.Sprintf("%s %s %s", quoteIdent(d, c.Column), string(c.Op), placeholder(d, argOffset))
	return sql, []any{c.Value}, argOffset + 1, nil
}

// InCondition represents "col" IN (?, ?, ...).
type InCondition struct {
	Column string
	Values []any
}

func (c InCondition) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if len(c.Values) == 0 {
		return "", nil, 0, fmt.Errorf("IN requires at least one value")
	}
	if len(c.Values) > MaxInValues {
		return "", nil, 0, fmt.Errorf("IN clause exceeds maximum values (%d), got %d", MaxInValues, len(c.Values))
	}
	if err := ValidColumnName(c.Column); err != nil {
		return "", nil, 0, fmt.Errorf("invalid column %q: %w", c.Column, err)
	}

	phs := make([]string, len(c.Values))
	for i := range c.Values {
		phs[i] = placeholder(d, argOffset+i)
	}
	sql := fmt.Sprintf("%s IN (%s)", quoteIdent(d, c.Column), strings.Join(phs, ", "))
	return sql, c.Values, argOffset + len(c.Values), nil
}

// BetweenCondition represents "col" BETWEEN ? AND ?.
type BetweenCondition struct {
	Column string
	Low    any
	High   any
}

func (c BetweenCondition) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if err := ValidColumnName(c.Column); err != nil {
		return "", nil, 0, fmt.Errorf("invalid column %q: %w", c.Column, err)
	}
	sql := fmt.Sprintf("%s BETWEEN %s AND %s",
		quoteIdent(d, c.Column),
		placeholder(d, argOffset),
		placeholder(d, argOffset+1),
	)
	return sql, []any{c.Low, c.High}, argOffset + 2, nil
}

// IsNullCondition represents "col" IS NULL.
type IsNullCondition struct {
	Column string
}

func (c IsNullCondition) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if err := ValidColumnName(c.Column); err != nil {
		return "", nil, 0, fmt.Errorf("invalid column %q: %w", c.Column, err)
	}
	return quoteIdent(d, c.Column) + " IS NULL", nil, argOffset, nil
}

// IsNotNullCondition represents "col" IS NOT NULL.
type IsNotNullCondition struct {
	Column string
}

func (c IsNotNullCondition) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if err := ValidColumnName(c.Column); err != nil {
		return "", nil, 0, fmt.Errorf("invalid column %q: %w", c.Column, err)
	}
	return quoteIdent(d, c.Column) + " IS NOT NULL", nil, argOffset, nil
}

// And represents a conjunction of conditions: (c1 AND c2 AND ...).
type And struct {
	Conditions []Condition
}

func (c And) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if len(c.Conditions) == 0 {
		return "", nil, 0, fmt.Errorf("AND requires at least one child condition")
	}

	deeper := BuildContext{
		CurrentDepth: ctx.CurrentDepth + 1,
		NodeCount:    ctx.NodeCount,
	}
	if err := deeper.checkDepth(); err != nil {
		return "", nil, 0, err
	}

	parts := make([]string, len(c.Conditions))
	var allArgs []any
	offset := argOffset

	for i, child := range c.Conditions {
		sql, args, nextOffset, err := child.Build(deeper, d, offset)
		if err != nil {
			return "", nil, 0, fmt.Errorf("AND child %d: %w", i, err)
		}
		parts[i] = sql
		allArgs = append(allArgs, args...)
		offset = nextOffset
	}

	return "(" + strings.Join(parts, " AND ") + ")", allArgs, offset, nil
}

// Or represents a disjunction of conditions: (c1 OR c2 OR ...).
type Or struct {
	Conditions []Condition
}

func (c Or) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if len(c.Conditions) == 0 {
		return "", nil, 0, fmt.Errorf("OR requires at least one child condition")
	}

	deeper := BuildContext{
		CurrentDepth: ctx.CurrentDepth + 1,
		NodeCount:    ctx.NodeCount,
	}
	if err := deeper.checkDepth(); err != nil {
		return "", nil, 0, err
	}

	parts := make([]string, len(c.Conditions))
	var allArgs []any
	offset := argOffset

	for i, child := range c.Conditions {
		sql, args, nextOffset, err := child.Build(deeper, d, offset)
		if err != nil {
			return "", nil, 0, fmt.Errorf("OR child %d: %w", i, err)
		}
		parts[i] = sql
		allArgs = append(allArgs, args...)
		offset = nextOffset
	}

	return "(" + strings.Join(parts, " OR ") + ")", allArgs, offset, nil
}

// Not represents a negation: NOT (condition).
type Not struct {
	Condition Condition // single condition; wrap in And/Or for multiple
}

func (c Not) Build(ctx BuildContext, d Dialect, argOffset int) (string, []any, int, error) {
	if err := ctx.incrementNode(); err != nil {
		return "", nil, 0, err
	}
	if c.Condition == nil {
		return "", nil, 0, fmt.Errorf("NOT requires a non-nil condition")
	}

	deeper := BuildContext{
		CurrentDepth: ctx.CurrentDepth + 1,
		NodeCount:    ctx.NodeCount,
	}
	if err := deeper.checkDepth(); err != nil {
		return "", nil, 0, err
	}

	sql, args, nextOffset, err := c.Condition.Build(deeper, d, argOffset)
	if err != nil {
		return "", nil, 0, fmt.Errorf("NOT: %w", err)
	}

	return "NOT (" + sql + ")", args, nextOffset, nil
}

// ValidateCondition walks the condition tree and returns an error if any
// safety limit is exceeded, any column name is invalid, or any operator
// is not in the allowlist. This is the authoritative pre-flight check
// before Build(). Build() also enforces limits defensively (belt-and-
// suspenders), but callers rely on ValidateCondition for user-facing
// error messages.
func ValidateCondition(c Condition) error {
	if c == nil {
		return fmt.Errorf("condition cannot be nil")
	}
	ctx := NewBuildContext()
	_, _, _, err := c.Build(ctx, DialectSQLite, 1)
	return err
}

// HasValueBinding returns true if the condition tree contains at least one
// leaf node that binds a parameterized value (Compare, InCondition, or
// BetweenCondition). Used by QUpdateFiltered and QDeleteFiltered to reject
// structurally valid but semantically vacuous conditions like
// IsNullCondition{Column: "id"} which would match all rows.
func HasValueBinding(c Condition) bool {
	switch v := c.(type) {
	case Compare:
		return true
	case InCondition:
		return true
	case BetweenCondition:
		return true
	case IsNullCondition:
		return false
	case IsNotNullCondition:
		return false
	case And:
		for _, child := range v.Conditions {
			if HasValueBinding(child) {
				return true
			}
		}
		return false
	case Or:
		for _, child := range v.Conditions {
			if HasValueBinding(child) {
				return true
			}
		}
		return false
	case Not:
		return HasValueBinding(v.Condition)
	default:
		return false
	}
}
