package tui

import "testing"

// ============================================================
// CellCount
// ============================================================

func TestGrid_CellCount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		grid Grid
		want int
	}{
		{
			name: "empty grid",
			grid: Grid{},
			want: 0,
		},
		{
			name: "single column single cell",
			grid: Grid{Columns: []GridColumn{
				{Span: 12, Cells: []GridCell{{Title: "A"}}},
			}},
			want: 1,
		},
		{
			name: "two columns multiple cells",
			grid: Grid{Columns: []GridColumn{
				{Span: 6, Cells: []GridCell{{Title: "A"}, {Title: "B"}}},
				{Span: 6, Cells: []GridCell{{Title: "C"}}},
			}},
			want: 3,
		},
		{
			name: "column with no cells",
			grid: Grid{Columns: []GridColumn{
				{Span: 6, Cells: []GridCell{{Title: "A"}}},
				{Span: 6, Cells: nil},
			}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.grid.CellCount()
			if got != tt.want {
				t.Errorf("CellCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ============================================================
// CellTitle
// ============================================================

func TestGrid_CellTitle(t *testing.T) {
	t.Parallel()
	grid := Grid{Columns: []GridColumn{
		{Span: 4, Cells: []GridCell{{Title: "Alpha"}, {Title: "Beta"}}},
		{Span: 4, Cells: []GridCell{{Title: "Gamma"}}},
		{Span: 4, Cells: []GridCell{{Title: "Delta"}}},
	}}

	tests := []struct {
		idx  int
		want string
	}{
		{0, "Alpha"},
		{1, "Beta"},
		{2, "Gamma"},
		{3, "Delta"},
		{4, ""},  // out of range
		{-1, ""}, // negative index
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			got := grid.CellTitle(tt.idx)
			if got != tt.want {
				t.Errorf("CellTitle(%d) = %q, want %q", tt.idx, got, tt.want)
			}
		})
	}
}

// ============================================================
// columnWidths
// ============================================================

func TestGrid_ColumnWidths(t *testing.T) {
	t.Parallel()

	t.Run("equal split", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{
			{Span: 6},
			{Span: 6},
		}}
		widths := grid.columnWidths(120)
		if widths[0] != 60 {
			t.Errorf("col 0 width = %d, want 60", widths[0])
		}
		if widths[1] != 60 {
			t.Errorf("col 1 width = %d, want 60", widths[1])
		}
	})

	t.Run("remainder goes to last column", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{
			{Span: 4},
			{Span: 4},
			{Span: 4},
		}}
		widths := grid.columnWidths(100)
		// 100 * 4/12 = 33 each, total 99, last gets +1
		total := 0
		for _, w := range widths {
			total += w
		}
		if total != 100 {
			t.Errorf("total width = %d, want 100", total)
		}
		if widths[2] != 34 {
			t.Errorf("last col width = %d, want 34 (33 + 1 remainder)", widths[2])
		}
	})

	t.Run("single column gets full width", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{{Span: 12}}}
		widths := grid.columnWidths(80)
		if widths[0] != 80 {
			t.Errorf("single col width = %d, want 80", widths[0])
		}
	})
}

// ============================================================
// cellHeights
// ============================================================

func TestGrid_CellHeights(t *testing.T) {
	t.Parallel()

	t.Run("equal heights", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{
			{Span: 12, Cells: []GridCell{
				{Height: 1},
				{Height: 1},
			}},
		}}
		heights := grid.cellHeights(0, 40)
		total := 0
		for _, h := range heights {
			total += h
		}
		if total != 40 {
			t.Errorf("total height = %d, want 40", total)
		}
	})

	t.Run("minimum height enforced", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{
			{Span: 12, Cells: []GridCell{
				{Height: 0.01}, // would compute to ~0
				{Height: 10},
			}},
		}}
		heights := grid.cellHeights(0, 10)
		if heights[0] < 3 {
			t.Errorf("cell height = %d, minimum is 3", heights[0])
		}
	})

	t.Run("zero height ratios default to equal", func(t *testing.T) {
		t.Parallel()
		grid := Grid{Columns: []GridColumn{
			{Span: 12, Cells: []GridCell{
				{Height: 0},
				{Height: 0},
			}},
		}}
		heights := grid.cellHeights(0, 20)
		total := 0
		for _, h := range heights {
			total += h
		}
		if total != 20 {
			t.Errorf("total height = %d, want 20", total)
		}
	})
}

// ============================================================
// CellInnerHeight
// ============================================================

func TestGrid_CellInnerHeight(t *testing.T) {
	t.Parallel()
	grid := Grid{Columns: []GridColumn{
		{Span: 6, Cells: []GridCell{
			{Title: "A", Height: 1},
		}},
		{Span: 6, Cells: []GridCell{
			{Title: "B", Height: 1},
		}},
	}}

	// Cell 0 is in column 0, cell 1 is in column 1
	h0 := grid.CellInnerHeight(0, 20)
	h1 := grid.CellInnerHeight(1, 20)

	// Both cells have height 1 in their columns, so both get full column height.
	// PanelInnerHeight subtracts 3 (border + title).
	if h0 != 17 {
		t.Errorf("CellInnerHeight(0, 20) = %d, want 17", h0)
	}
	if h1 != 17 {
		t.Errorf("CellInnerHeight(1, 20) = %d, want 17", h1)
	}

	// Out of range falls back to PanelInnerHeight(totalHeight)
	hOOB := grid.CellInnerHeight(99, 20)
	if hOOB != 17 {
		t.Errorf("CellInnerHeight(99, 20) = %d, want 17", hOOB)
	}
}
