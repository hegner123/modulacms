package plugin

import (
	"strings"
	"testing"
)

func TestRenderList(t *testing.T) {
	prim := &ListPrimitive{
		Items: []ListItem{
			{Label: "First", ID: "1"},
			{Label: "Second", ID: "2"},
			{Label: "Third", ID: "3", Faint: true},
		},
		Cursor: 1,
	}

	result := RenderPrimitive(prim, 40, 10, false)
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "First") {
		t.Errorf("line 0 missing 'First': %q", lines[0])
	}
	if !strings.Contains(lines[1], "->") {
		t.Errorf("line 1 missing cursor '->': %q", lines[1])
	}
	if !strings.Contains(lines[1], "Second") {
		t.Errorf("line 1 missing 'Second': %q", lines[1])
	}
}

func TestRenderList_Empty(t *testing.T) {
	prim := &ListPrimitive{EmptyText: "nothing here"}
	result := RenderPrimitive(prim, 40, 10, false)
	if !strings.Contains(result, "nothing here") {
		t.Errorf("expected empty text, got %q", result)
	}
}

func TestRenderDetail(t *testing.T) {
	prim := &DetailPrimitive{
		Fields: []DetailField{
			{Label: "Name", Value: "Test Item"},
			{Label: "Status", Value: "active"},
		},
	}

	result := RenderPrimitive(prim, 40, 10, false)
	if !strings.Contains(result, "Name") {
		t.Errorf("missing 'Name': %q", result)
	}
	if !strings.Contains(result, "Test Item") {
		t.Errorf("missing 'Test Item': %q", result)
	}
}

func TestRenderText(t *testing.T) {
	prim := &TextPrimitive{
		Lines: []TextLine{
			{Text: "Hello world"},
			{Text: ""},
			{Text: "Bold line", Bold: true},
		},
	}

	result := RenderPrimitive(prim, 40, 10, false)
	if !strings.Contains(result, "Hello world") {
		t.Errorf("missing 'Hello world': %q", result)
	}
}

func TestRenderTable(t *testing.T) {
	prim := &TablePrimitive{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"Item A", "active"},
			{"Item B", "draft"},
		},
		Cursor: 0,
	}

	result := RenderPrimitive(prim, 60, 10, false)
	if !strings.Contains(result, "Name") {
		t.Errorf("missing header 'Name': %q", result)
	}
	if !strings.Contains(result, "->") {
		t.Errorf("missing cursor '->': %q", result)
	}
	if !strings.Contains(result, "Item A") {
		t.Errorf("missing 'Item A': %q", result)
	}
}

func TestRenderInput(t *testing.T) {
	prim := &InputPrimitive{
		ID:          "search",
		Value:       "hello",
		Placeholder: "Type...",
		Focused:     true,
	}

	result := RenderPrimitive(prim, 40, 1, false)
	if !strings.HasPrefix(result, "> ") {
		t.Errorf("expected focused prefix '> ', got %q", result)
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("missing value 'hello': %q", result)
	}
}

func TestRenderInput_Placeholder(t *testing.T) {
	prim := &InputPrimitive{
		ID:          "search",
		Value:       "",
		Placeholder: "Type...",
		Focused:     false,
	}

	result := RenderPrimitive(prim, 40, 1, false)
	if !strings.Contains(result, "Type...") {
		t.Errorf("missing placeholder: %q", result)
	}
}

func TestRenderSelect(t *testing.T) {
	prim := &SelectPrimitive{
		ID: "filter",
		Options: []SelectOption{
			{Label: "All", Value: ""},
			{Label: "Active", Value: "active"},
		},
		Selected: 1,
	}

	result := RenderPrimitive(prim, 40, 1, false)
	if !strings.Contains(result, "Active") {
		t.Errorf("missing selected option 'Active': %q", result)
	}
}

func TestRenderTree(t *testing.T) {
	prim := &TreePrimitive{
		Nodes: []TreeNode{
			{Label: "Root", ID: "r1", Expanded: true, Children: []TreeNode{
				{Label: "Child A", ID: "c1"},
				{Label: "Child B", ID: "c2"},
			}},
		},
		Cursor: 1, // Child A
	}

	result := RenderPrimitive(prim, 40, 10, false)
	if !strings.Contains(result, "Root") {
		t.Errorf("missing 'Root': %q", result)
	}
	if !strings.Contains(result, "Child A") {
		t.Errorf("missing 'Child A': %q", result)
	}
	if !strings.Contains(result, "->") {
		t.Errorf("missing cursor: %q", result)
	}
}

func TestRenderProgress(t *testing.T) {
	prim := &ProgressPrimitive{
		Value: 0.5,
		Label: "Loading",
	}

	result := RenderPrimitive(prim, 40, 1, false)
	if !strings.Contains(result, "█") {
		t.Errorf("missing filled bar: %q", result)
	}
	if !strings.Contains(result, "50%") {
		t.Errorf("missing percentage: %q", result)
	}
	if !strings.Contains(result, "Loading") {
		t.Errorf("missing label: %q", result)
	}
}

func TestLayoutToCells(t *testing.T) {
	layout := &PluginLayout{
		Columns: []PluginColumn{
			{Span: 3, Cells: []PluginCell{
				{Title: "List", Height: 1.0, Content: &ListPrimitive{
					Items: []ListItem{{Label: "A", ID: "1"}}, Cursor: 0,
				}},
			}},
			{Span: 9, Cells: []PluginCell{
				{Title: "Detail", Height: 0.6, Content: &DetailPrimitive{
					Fields: []DetailField{{Label: "Name", Value: "Test"}},
				}},
				{Title: "Info", Height: 0.4, Content: &TextPrimitive{
					Lines: []TextLine{{Text: "Info"}},
				}},
			}},
		},
		Hints: []PluginHint{{Key: "n", Label: "new"}},
	}

	cells, gridDef := LayoutToCells(layout, 120, 40, 0)
	if len(cells) != 3 {
		t.Fatalf("expected 3 cells, got %d", len(cells))
	}
	if len(gridDef.Columns) != 2 {
		t.Fatalf("expected 2 grid columns, got %d", len(gridDef.Columns))
	}
	if gridDef.Columns[0].Span != 3 {
		t.Errorf("expected span 3, got %d", gridDef.Columns[0].Span)
	}
	if gridDef.Columns[1].Cells[0].Title != "Detail" {
		t.Errorf("expected title 'Detail', got %q", gridDef.Columns[1].Cells[0].Title)
	}
	if cells[0].Content == "" {
		t.Error("cell 0 content is empty")
	}
}
