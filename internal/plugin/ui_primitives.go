package plugin

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Validation limits for plugin UI yields.
const (
	maxListItems   = 10000
	maxTableRows   = 10000
	maxStringLen   = 10240 // 10KB
	maxTreeDepth   = 20
	maxColumnCount = 12
)

// PluginPrimitive is the interface for all UI primitive types.
type PluginPrimitive interface {
	PrimitiveType() string
}

// ListPrimitive represents a vertical item list with cursor.
type ListPrimitive struct {
	Items     []ListItem
	Cursor    int
	EmptyText string
}

func (p *ListPrimitive) PrimitiveType() string { return "list" }

// ListItem is a single item in a list primitive.
type ListItem struct {
	Label string
	ID    string
	Faint bool
	Bold  bool
}

// DetailPrimitive represents a key-value pair display.
type DetailPrimitive struct {
	Title  string
	Fields []DetailField
}

func (p *DetailPrimitive) PrimitiveType() string { return "detail" }

// DetailField is a single field in a detail primitive.
type DetailField struct {
	Label string
	Value string
	Faint bool
}

// TextPrimitive represents a styled text block.
type TextPrimitive struct {
	Lines []TextLine
}

func (p *TextPrimitive) PrimitiveType() string { return "text" }

// TextLine is a single line in a text primitive.
type TextLine struct {
	Text   string
	Bold   bool
	Faint  bool
	Accent bool
	Style  *TextStyle
}

// TextStyle holds optional styling for a text line.
type TextStyle struct {
	Fg string
	Bg string
}

// TablePrimitive represents a table with headers and rows.
type TablePrimitive struct {
	Headers []string
	Rows    [][]string
	Cursor  int
}

func (p *TablePrimitive) PrimitiveType() string { return "table" }

// InputPrimitive represents a text input field.
type InputPrimitive struct {
	ID          string
	Placeholder string
	Value       string
	Focused     bool
	CharLimit   int
}

func (p *InputPrimitive) PrimitiveType() string { return "input" }

// SelectPrimitive represents an option selector.
type SelectPrimitive struct {
	ID       string
	Options  []SelectOption
	Selected int
	Focused  bool
}

func (p *SelectPrimitive) PrimitiveType() string { return "select" }

// SelectOption is a single option in a select primitive.
type SelectOption struct {
	Label string
	Value string
}

// TreePrimitive represents a hierarchical expandable tree.
type TreePrimitive struct {
	Nodes  []TreeNode
	Cursor int
}

func (p *TreePrimitive) PrimitiveType() string { return "tree" }

// TreeNode is a single node in a tree primitive.
type TreeNode struct {
	Label    string
	ID       string
	Expanded bool
	Children []TreeNode
}

// ProgressPrimitive represents a progress bar.
type ProgressPrimitive struct {
	Value float64 // 0.0 to 1.0
	Label string
}

func (p *ProgressPrimitive) PrimitiveType() string { return "progress" }

// ParseLayout parses a grid layout from a Lua table.
func ParseLayout(tbl *lua.LTable) (*PluginLayout, error) {
	layout := &PluginLayout{}

	// Parse columns.
	colsVal := tbl.RawGetString("columns")
	colsTbl, ok := colsVal.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("grid missing 'columns' table")
	}

	colCount := colsTbl.MaxN()
	if colCount > maxColumnCount {
		return nil, fmt.Errorf("grid has %d columns (max %d)", colCount, maxColumnCount)
	}

	for i := 1; i <= colCount; i++ {
		colVal := colsTbl.RawGetInt(i)
		colTbl, ok := colVal.(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("column %d is not a table", i)
		}
		col, err := parseColumn(colTbl)
		if err != nil {
			return nil, fmt.Errorf("column %d: %w", i, err)
		}
		layout.Columns = append(layout.Columns, col)
	}

	// Parse hints (optional).
	hintsVal := tbl.RawGetString("hints")
	if hintsTbl, ok := hintsVal.(*lua.LTable); ok {
		hintCount := hintsTbl.MaxN()
		for i := 1; i <= hintCount; i++ {
			hintVal := hintsTbl.RawGetInt(i)
			hintTbl, ok := hintVal.(*lua.LTable)
			if !ok {
				continue
			}
			hint := PluginHint{
				Key:   luaRawString(hintTbl, "key"),
				Label: luaRawString(hintTbl, "label"),
			}
			if hint.Key != "" {
				layout.Hints = append(layout.Hints, hint)
			}
		}
	}

	return layout, nil
}

func parseColumn(tbl *lua.LTable) (PluginColumn, error) {
	col := PluginColumn{}

	spanVal := tbl.RawGetString("span")
	if num, ok := spanVal.(lua.LNumber); ok {
		col.Span = int(num)
	}
	if col.Span <= 0 {
		col.Span = 1
	}

	cellsVal := tbl.RawGetString("cells")
	cellsTbl, ok := cellsVal.(*lua.LTable)
	if !ok {
		return col, fmt.Errorf("column missing 'cells' table")
	}

	cellCount := cellsTbl.MaxN()
	for i := 1; i <= cellCount; i++ {
		cellVal := cellsTbl.RawGetInt(i)
		cellTbl, ok := cellVal.(*lua.LTable)
		if !ok {
			return col, fmt.Errorf("cell %d is not a table", i)
		}
		cell, err := parseCell(cellTbl)
		if err != nil {
			return col, fmt.Errorf("cell %d: %w", i, err)
		}
		col.Cells = append(col.Cells, cell)
	}

	return col, nil
}

func parseCell(tbl *lua.LTable) (PluginCell, error) {
	cell := PluginCell{
		Title:  luaRawString(tbl, "title"),
		Height: 1.0,
	}

	heightVal := tbl.RawGetString("height")
	if num, ok := heightVal.(lua.LNumber); ok {
		cell.Height = float64(num)
	}
	if cell.Height <= 0 {
		cell.Height = 1.0
	}

	contentVal := tbl.RawGetString("content")
	contentTbl, ok := contentVal.(*lua.LTable)
	if !ok {
		return cell, fmt.Errorf("cell missing 'content' table")
	}

	prim, err := ParsePrimitive(contentTbl)
	if err != nil {
		return cell, fmt.Errorf("cell content: %w", err)
	}
	cell.Content = prim

	return cell, nil
}

// ParsePrimitive parses a single UI primitive from a Lua table.
func ParsePrimitive(tbl *lua.LTable) (PluginPrimitive, error) {
	typeVal := tbl.RawGetString("type")
	typeStr, ok := typeVal.(lua.LString)
	if !ok {
		return nil, fmt.Errorf("primitive missing 'type' field")
	}

	switch string(typeStr) {
	case "list":
		return parseListPrimitive(tbl)
	case "detail":
		return parseDetailPrimitive(tbl)
	case "text":
		return parseTextPrimitive(tbl)
	case "table":
		return parseTablePrimitive(tbl)
	case "input":
		return parseInputPrimitive(tbl)
	case "select":
		return parseSelectPrimitive(tbl)
	case "tree":
		return parseTreePrimitive(tbl)
	case "progress":
		return parseProgressPrimitive(tbl)
	default:
		return nil, fmt.Errorf("unknown primitive type %q", string(typeStr))
	}
}

func parseListPrimitive(tbl *lua.LTable) (*ListPrimitive, error) {
	prim := &ListPrimitive{
		EmptyText: luaRawString(tbl, "empty_text"),
	}

	cursorVal := tbl.RawGetString("cursor")
	if num, ok := cursorVal.(lua.LNumber); ok {
		prim.Cursor = int(num)
	}

	itemsVal := tbl.RawGetString("items")
	if itemsTbl, ok := itemsVal.(*lua.LTable); ok {
		itemCount := itemsTbl.MaxN()
		if itemCount > maxListItems {
			return nil, fmt.Errorf("list has %d items (max %d)", itemCount, maxListItems)
		}
		for i := 1; i <= itemCount; i++ {
			itemVal := itemsTbl.RawGetInt(i)
			itemTbl, ok := itemVal.(*lua.LTable)
			if !ok {
				continue
			}
			item := ListItem{
				Label: luaRawString(itemTbl, "label"),
				ID:    luaRawString(itemTbl, "id"),
				Faint: luaRawBool(itemTbl, "faint"),
				Bold:  luaRawBool(itemTbl, "bold"),
			}
			if len(item.Label) > maxStringLen {
				item.Label = item.Label[:maxStringLen]
			}
			prim.Items = append(prim.Items, item)
		}
	}

	return prim, nil
}

func parseDetailPrimitive(tbl *lua.LTable) (*DetailPrimitive, error) {
	prim := &DetailPrimitive{
		Title: luaRawString(tbl, "title"),
	}

	fieldsVal := tbl.RawGetString("fields")
	if fieldsTbl, ok := fieldsVal.(*lua.LTable); ok {
		fieldCount := fieldsTbl.MaxN()
		for i := 1; i <= fieldCount; i++ {
			fieldVal := fieldsTbl.RawGetInt(i)
			fieldTbl, ok := fieldVal.(*lua.LTable)
			if !ok {
				continue
			}
			field := DetailField{
				Label: luaRawString(fieldTbl, "label"),
				Value: luaRawString(fieldTbl, "value"),
				Faint: luaRawBool(fieldTbl, "faint"),
			}
			if len(field.Value) > maxStringLen {
				field.Value = field.Value[:maxStringLen]
			}
			prim.Fields = append(prim.Fields, field)
		}
	}

	return prim, nil
}

func parseTextPrimitive(tbl *lua.LTable) (*TextPrimitive, error) {
	prim := &TextPrimitive{}

	linesVal := tbl.RawGetString("lines")
	if linesTbl, ok := linesVal.(*lua.LTable); ok {
		lineCount := linesTbl.MaxN()
		for i := 1; i <= lineCount; i++ {
			lineVal := linesTbl.RawGetInt(i)
			var line TextLine
			switch v := lineVal.(type) {
			case lua.LString:
				text := string(v)
				if len(text) > maxStringLen {
					text = text[:maxStringLen]
				}
				line = TextLine{Text: text}
			case *lua.LTable:
				line = TextLine{
					Text:   luaRawString(v, "text"),
					Bold:   luaRawBool(v, "bold"),
					Faint:  luaRawBool(v, "faint"),
					Accent: luaRawBool(v, "accent"),
				}
				if len(line.Text) > maxStringLen {
					line.Text = line.Text[:maxStringLen]
				}
				styleVal := v.RawGetString("style")
				if styleTbl, ok := styleVal.(*lua.LTable); ok {
					line.Style = &TextStyle{
						Fg: luaRawString(styleTbl, "fg"),
						Bg: luaRawString(styleTbl, "bg"),
					}
				}
			default:
				line = TextLine{Text: lineVal.String()}
			}
			prim.Lines = append(prim.Lines, line)
		}
	}

	return prim, nil
}

func parseTablePrimitive(tbl *lua.LTable) (*TablePrimitive, error) {
	prim := &TablePrimitive{}

	cursorVal := tbl.RawGetString("cursor")
	if num, ok := cursorVal.(lua.LNumber); ok {
		prim.Cursor = int(num)
	}

	// Parse headers.
	headersVal := tbl.RawGetString("headers")
	if headersTbl, ok := headersVal.(*lua.LTable); ok {
		headerCount := headersTbl.MaxN()
		for i := 1; i <= headerCount; i++ {
			hdrVal := headersTbl.RawGetInt(i)
			if s, ok := hdrVal.(lua.LString); ok {
				prim.Headers = append(prim.Headers, string(s))
			}
		}
	}

	// Parse rows.
	rowsVal := tbl.RawGetString("rows")
	if rowsTbl, ok := rowsVal.(*lua.LTable); ok {
		rowCount := rowsTbl.MaxN()
		if rowCount > maxTableRows {
			return nil, fmt.Errorf("table has %d rows (max %d)", rowCount, maxTableRows)
		}
		for i := 1; i <= rowCount; i++ {
			rowVal := rowsTbl.RawGetInt(i)
			rowTbl, ok := rowVal.(*lua.LTable)
			if !ok {
				continue
			}
			var row []string
			colCount := rowTbl.MaxN()
			for j := 1; j <= colCount; j++ {
				cellVal := rowTbl.RawGetInt(j)
				s := cellVal.String()
				if len(s) > maxStringLen {
					s = s[:maxStringLen]
				}
				row = append(row, s)
			}
			prim.Rows = append(prim.Rows, row)
		}
	}

	return prim, nil
}

func parseInputPrimitive(tbl *lua.LTable) (*InputPrimitive, error) {
	prim := &InputPrimitive{
		ID:          luaRawString(tbl, "id"),
		Placeholder: luaRawString(tbl, "placeholder"),
		Value:       luaRawString(tbl, "value"),
		Focused:     luaRawBool(tbl, "focused"),
	}

	limitVal := tbl.RawGetString("char_limit")
	if num, ok := limitVal.(lua.LNumber); ok {
		prim.CharLimit = int(num)
	}

	return prim, nil
}

func parseSelectPrimitive(tbl *lua.LTable) (*SelectPrimitive, error) {
	prim := &SelectPrimitive{
		ID:      luaRawString(tbl, "id"),
		Focused: luaRawBool(tbl, "focused"),
	}

	selectedVal := tbl.RawGetString("selected")
	if num, ok := selectedVal.(lua.LNumber); ok {
		prim.Selected = int(num)
	}

	optionsVal := tbl.RawGetString("options")
	if optsTbl, ok := optionsVal.(*lua.LTable); ok {
		optCount := optsTbl.MaxN()
		for i := 1; i <= optCount; i++ {
			optVal := optsTbl.RawGetInt(i)
			optTbl, ok := optVal.(*lua.LTable)
			if !ok {
				continue
			}
			opt := SelectOption{
				Label: luaRawString(optTbl, "label"),
				Value: luaRawString(optTbl, "value"),
			}
			prim.Options = append(prim.Options, opt)
		}
	}

	return prim, nil
}

func parseTreePrimitive(tbl *lua.LTable) (*TreePrimitive, error) {
	prim := &TreePrimitive{}

	cursorVal := tbl.RawGetString("cursor")
	if num, ok := cursorVal.(lua.LNumber); ok {
		prim.Cursor = int(num)
	}

	nodesVal := tbl.RawGetString("nodes")
	if nodesTbl, ok := nodesVal.(*lua.LTable); ok {
		nodes, err := parseTreeNodes(nodesTbl, 0)
		if err != nil {
			return nil, err
		}
		prim.Nodes = nodes
	}

	return prim, nil
}

func parseTreeNodes(tbl *lua.LTable, depth int) ([]TreeNode, error) {
	if depth > maxTreeDepth {
		return nil, fmt.Errorf("tree depth exceeds %d levels", maxTreeDepth)
	}

	var nodes []TreeNode
	nodeCount := tbl.MaxN()
	for i := 1; i <= nodeCount; i++ {
		nodeVal := tbl.RawGetInt(i)
		nodeTbl, ok := nodeVal.(*lua.LTable)
		if !ok {
			continue
		}

		node := TreeNode{
			Label:    luaRawString(nodeTbl, "label"),
			ID:       luaRawString(nodeTbl, "id"),
			Expanded: luaRawBool(nodeTbl, "expanded"),
		}

		childrenVal := nodeTbl.RawGetString("children")
		if childrenTbl, ok := childrenVal.(*lua.LTable); ok {
			children, err := parseTreeNodes(childrenTbl, depth+1)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func parseProgressPrimitive(tbl *lua.LTable) (*ProgressPrimitive, error) {
	prim := &ProgressPrimitive{
		Label: luaRawString(tbl, "label"),
	}

	valueVal := tbl.RawGetString("value")
	if num, ok := valueVal.(lua.LNumber); ok {
		prim.Value = float64(num)
		if prim.Value < 0 {
			prim.Value = 0
		}
		if prim.Value > 1 {
			prim.Value = 1
		}
	}

	return prim, nil
}

// luaRawString reads a string field from a Lua table using raw access.
func luaRawString(tbl *lua.LTable, field string) string {
	val := tbl.RawGetString(field)
	if s, ok := val.(lua.LString); ok {
		return string(s)
	}
	return ""
}

// luaRawBool reads a boolean field from a Lua table using raw access.
func luaRawBool(tbl *lua.LTable, field string) bool {
	val := tbl.RawGetString(field)
	if b, ok := val.(lua.LBool); ok {
		return bool(b)
	}
	return false
}
