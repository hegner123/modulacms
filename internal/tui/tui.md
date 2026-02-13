# tui

The tui package implements a Bubbletea-based terminal user interface for ModulaCMS. It provides an SSH-accessible TUI with three-panel layout for managing content trees, content fields, and routing. The package follows the Elm Architecture pattern with separate model, view, and update components.

## Overview

This package integrates with the Charmbracelet ecosystem to provide a full-featured TUI that runs over SSH. The main components are the Model struct that holds application state, a View function that renders the UI, and an Update function that handles user input and state transitions. The TUI displays a header with action buttons, a three-panel body area for content management, and a status bar with keyboard hints.

The layout uses three vertical panels with percentage-based widths. The Tree panel on the left shows content hierarchy, the Content panel in the center displays selected node fields, and the Route panel on the right shows routing information. Users navigate between panels using Tab and Shift+Tab keys.

## Types

### Model

Model is the top-level Bubbletea model for the TUI. It implements the tea.Model interface and contains all application state including configuration, terminal dimensions, current time, verbosity flag, and the currently focused panel.

The Model struct has six fields: Config holds the application configuration pointer, Width and Height track terminal dimensions in columns and rows, Term stores the terminal type string, Time holds the current timestamp, Verbose is a boolean flag for debug output, and Focus indicates which panel has keyboard focus.

### FocusPanel

FocusPanel is an integer type that identifies which panel currently has keyboard focus. It has three constant values: TreePanel with value zero for the content tree panel, ContentPanel with value one for the fields panel, and RoutePanel with value two for the routing panel.

The FocusPanel type provides a String method that returns the display name of the focused panel. It returns "Tree" for TreePanel, "Content" for ContentPanel, "Route" for RoutePanel, or "Unknown" for any other value.

### Panel

Panel represents a bordered UI section with a title. It is used to render each of the three main content areas in the TUI body. The struct contains five fields: Title is the panel heading text, Width and Height define dimensions in columns and rows, Content holds the text to display inside the panel, and Focused is a boolean that determines border color.

### Overlay

Overlay represents a positioned content rectangle to render on top of a base layer. It is used for compositing layered UI elements. The struct has five fields: Content is the pre-rendered content string for this layer, X is the left column origin in zero-based coordinates, Y is the top row origin, Width is the overlay width in columns, and Height is the overlay height in rows.

## Functions

### InitialModel

InitialModel creates a Model with the given config wired in and no initial command. It accepts a verbose boolean pointer and a Config pointer, returning a Model and error. If the verbose pointer is nil, verbosity defaults to false. The Focus field is initialized to TreePanel.

The function signature is: func InitialModel(v *bool, c *config.Config) (Model, error). It returns a configured Model ready to be passed to tea.NewProgram. The error return value is currently always nil but follows the pattern for future initialization errors.

### Run

Run creates and runs a tea.Program with alt screen enabled. It accepts a Model pointer and returns a tea.Program pointer and a boolean indicating normal exit. The function creates a new Bubbletea program with alternate screen mode, runs it until completion, and logs any fatal errors using the utility logger.

The function signature is: func Run(m *Model) (*tea.Program, bool). The boolean return value is currently always false. The alternate screen mode ensures the TUI does not pollute the terminal scrollback buffer.

### TuiMiddleware

TuiMiddleware returns a wish.Middleware that launches the TUI for SSH sessions. It accepts a verbose boolean pointer and a Config pointer, returning a configured middleware function. The middleware creates a new Bubbletea program for each SSH connection with a one-second time ticker goroutine.

The teaHandler function inside the middleware extracts the PTY dimensions and terminal type from the SSH session, initializes the Model with session-specific values, and returns a configured tea.Program. The middleware uses termenv.ANSI256 color profile and enables alternate screen mode for each session.

## Methods

### Init

Init is the Bubbletea initialization method for Model. It returns a tea.Cmd value which is always nil because no initial async work is needed. This method satisfies the tea.Model interface requirement.

The method signature is: func (m Model) Init() tea.Cmd. It is called once when the Bubbletea program starts, before the first Update call.

### Update

Update handles messages for the TUI model. It accepts a tea.Msg and returns a tea.Model and tea.Cmd. The method processes KeyMsg events for quit commands and tab navigation, and WindowSizeMsg events to track terminal dimensions.

For KeyMsg events, the method quits on "q" or "ctrl+c", cycles focus forward on "tab" using modulo three arithmetic, and cycles focus backward on "shift+tab" using modulo three arithmetic with an offset of two. For WindowSizeMsg events, the method updates Width and Height fields to match the new terminal dimensions.

The method signature is: func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd). It returns the updated model and any command to execute. Most updates return nil for the command value.

### View

View renders the TUI. It returns a string containing the complete terminal output. The method first checks if Width or Height is zero and returns "Loading..." if dimensions are not yet available. Otherwise it renders the header, status bar, and three-panel body.

The body height is calculated as total height minus header height minus status bar height, with a minimum of three rows. Column widths are computed as percentages: left panel is one quarter, center panel is one half, and right panel is the remainder.

Three Panel structs are created with titles "Tree", "Content", and "Route", populated with placeholder content and appropriate dimensions. The Focused field is set based on the model's Focus value. Panels are rendered and joined horizontally, then combined vertically with header and status bar.

The method signature is: func (m Model) View() string. The returned string uses ANSI escape codes for styling and layout.

### Render

Render is a method on Panel that draws the panel as a bordered box with a title bar. It returns a string containing the rendered panel output. The method sets border color based on the Focused field, applies bold accent styling to the title, calculates inner dimensions accounting for borders, pads or truncates content, and renders the final box with lipgloss.

Border color is the default tertiary color when unfocused and the accent color when focused. Inner width subtracts two characters for left and right borders. Inner height subtracts three rows for top border, bottom border, and title row. Content is processed by padContent to fill the inner area exactly.

The method signature is: func (p Panel) Render() string. It uses lipgloss RoundedBorder style and returns a multi-line string with ANSI styling.

### String

String is a method on FocusPanel that returns the display name of the focused panel. It returns "Tree" for TreePanel constant, "Content" for ContentPanel constant, "Route" for RoutePanel constant, or "Unknown" for any other value.

The method signature is: func (f FocusPanel) String() string. It is used in status bar rendering to show which panel has focus.

### Composite

Composite renders the overlay on top of the base string buffer. It accepts a base string and an Overlay struct, returning a string with the overlay composited. Rows outside the overlay's Y range pass through unchanged. Rows inside the overlay's Y range have columns X through X+Width replaced with the corresponding overlay line.

The function splits base and overlay content into lines, extends the base if the overlay extends beyond it, and iterates through all base lines. For each line in the overlay's Y range, it splices the overlay content using spliceLine. Lines outside the range are unchanged.

The function signature is: func Composite(base string, overlay Overlay) string. It enables modal dialogs and popups to be rendered over the main UI without complex state management.

## Internal Functions

### renderHeader

The renderHeader function renders the top action bar with app title and action buttons. It accepts a width integer and returns a string. The function creates a bold accent-styled title "ModulaCMS" and five action buttons labeled New, Save, Copy, Duplicate, and Export. Buttons are joined horizontally and combined with the title, then placed in a container with bottom border.

### renderStatusBar

The renderStatusBar function renders the bottom status bar with focus indicator and key hints. It accepts a Model and returns a string. The bar displays the focused panel name in brackets on the left and keyboard hints "tab: switch panel q: quit" on the right. A spacer fills the gap between left and right sections to align hints to the right edge.

### padContent

The padContent function ensures content fills exactly the given width and height. It accepts a content string, width integer, and height integer, returning a padded string. Lines are added or removed to match height exactly. Each line is truncated to width using rune-aware counting to handle multibyte characters correctly.

### spliceLine

The spliceLine function replaces columns X through X+Width of base with top content. It accepts base string, top string, x integer, and width integer, returning a string. The function uses ANSI-aware truncation to split the base into left, middle, and right sections. The left section contains base content up to column x, the middle contains top content padded to exactly width columns, and the right contains base content after column x+width.
