---@meta

--- TUI primitives module. Pure helper constructors for building terminal UI
--- screens via coroutine-based rendering.
---
--- Usage:
--- ```lua
--- coroutine.yield(tui.grid({
---     tui.column(3, { tui.cell("List", 1.0, tui.list(items, cursor)) }),
---     tui.column(9, { tui.cell("Detail", 0.6, tui.detail(fields)) }),
--- }))
--- ```
---@class tui
tui = {}

---@class tui.Grid
---@field type "grid"
---@field columns tui.Column[]
---@field hints? tui.Hint[]

---@class tui.Column
---@field span integer Column width in grid units.
---@field cells tui.Cell[]

---@class tui.Cell
---@field title string Cell title.
---@field height number Relative height weight.
---@field content tui.Primitive

---@class tui.Hint
---@field key string Keyboard key (e.g., `"enter"`).
---@field label string Description of what the key does.

---@class tui.ListItem
---@field label string Display text.
---@field id? string Item identifier.
---@field faint? boolean Dim styling.
---@field bold? boolean Bold styling.

---@class tui.DetailField
---@field label string Field label.
---@field value string Field value.
---@field faint? boolean Dim styling.

---@class tui.TextLine
---@field text string Line content.
---@field bold? boolean
---@field faint? boolean
---@field accent? boolean
---@field style? tui.TextStyle

---@class tui.TextStyle
---@field fg? string Foreground color.
---@field bg? string Background color.

---@class tui.SelectOption
---@field label string Display text.
---@field value string Option value.

---@class tui.TreeNode
---@field label string Node display text.
---@field id? string Node identifier.
---@field expanded? boolean Whether children are visible.
---@field children? tui.TreeNode[]

--- Alias for any primitive content type.
---@alias tui.Primitive tui.Grid|table

--- Build a grid layout.
---@param columns tui.Column[] Array of columns.
---@param hints? tui.Hint[] Keyboard hint bar entries.
---@return tui.Grid
function tui.grid(columns, hints) end

--- Build a column.
---@param span integer Column width in grid units.
---@param cells tui.Cell[] Array of cells.
---@return tui.Column
function tui.column(span, cells) end

--- Build a cell.
---@param title string Cell title.
---@param height number Relative height weight.
---@param content tui.Primitive Cell content primitive.
---@return tui.Cell
function tui.cell(title, height, content) end

--- Build a list primitive.
---@param items tui.ListItem[] Array of list items.
---@param cursor? integer Selected index (default 0).
---@return table
function tui.list(items, cursor) end

--- Build a detail (key-value pairs) primitive.
---@param fields tui.DetailField[] Array of label-value fields.
---@return table
function tui.detail(fields) end

--- Build a text block primitive.
--- Items can be plain strings or `tui.TextLine` tables.
---@param lines (string|tui.TextLine)[] Array of text lines.
---@return table
function tui.text(lines) end

--- Build a table primitive.
---@param headers string[] Column headers.
---@param rows string[][] Array of row arrays.
---@param cursor? integer Selected row index (default 0).
---@return table
function tui.table(headers, rows, cursor) end

--- Build a text input primitive.
---@param id string Input identifier.
---@param value? string Current value (default `""`).
---@param placeholder? string Placeholder text (default `""`).
---@return table
function tui.input(id, value, placeholder) end

--- Build a select/dropdown primitive.
--- Named `select_field` because `select` is a Lua reserved word.
---@param id string Select identifier.
---@param options tui.SelectOption[] Available options.
---@param selected? integer Selected index (default 0).
---@return table
function tui.select_field(id, options, selected) end

--- Build a tree primitive.
---@param nodes tui.TreeNode[] Root-level tree nodes.
---@param cursor? integer Selected node index (default 0).
---@return table
function tui.tree(nodes, cursor) end

--- Build a progress bar primitive.
---@param value number Progress value from 0.0 to 1.0.
---@param label? string Label text (default `""`).
---@return table
function tui.progress(value, label) end
