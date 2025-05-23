# ModulaCMS CLI Package

This package implements the terminal user interface (TUI) for ModulaCMS using the [Charm](https://charm.sh/) libraries (bubbletea, lipgloss, and huh).

## Architecture

The CLI package follows the Model-View-Update (MVU) architecture pattern:

- **Model**: Central data store containing application state
- **Update**: Processes messages and updates the model
- **View**: Renders the current state to the terminal

## Core Components

### Model Structure
- `Model` struct as central state container
- Component-based approach with specialized components:
  - `ThemeComponent`: Styling and theming
  - `NavigationComponent`: Page navigation and history
  - `TableComponent`: Database table rendering
  - `FormComponent`: Form generation and state
  - `ContentComponent`: Content display

### Navigation System
- Hierarchical page structure with parent-child relationships
- `PageIndex` identifier and specific `Controller` interface for each page
- Navigation history tracking
- Menu system for application navigation

### Controller Pattern
- `CliInterface` type for page-specific functionality
- Specialized controllers (readInterface, updateInterface, etc.)
- Context-based update logic separation

### Rendering System
- Terminal styling with lipgloss
- Flexible layouts with title, header, body, and footer sections
- ASCII art titles from embedded files
- Terminal size responsiveness

### Dialog System
- Modal dialogs overlaying the main UI
- OK/Cancel button functionality
- Confirmation, alert, and error messages

### Database Operations
- CRUD operations for database tables
- Dynamic table fetching and rendering
- Automatic form generation from table structure
- Form validation

## Key Abstractions

### Page
- Distinct UI view with navigation metadata
- Associated controller interface

### Focus System
- Input focus management via `FocusKey` type
- Controls keyboard input destination

### Application State
- Status tracking with `ApplicationState`
- Visual status representation

## Implementation Patterns

### TEA (The Elm Architecture)
- Event-driven with message passing
- Initialize → Update → View pattern
- Pure functional updates

### Command Pattern
- Asynchronous operations with tea.Cmd
- Message-based command response processing
- Non-blocking UI

### Component-Based Architecture
- Modular, composable components
- Encapsulated functionality and state
- Composition over inheritance

## Message Types
- Event-specific message types
- Dialog messages
- Data-fetching messages

## File Structure
- `model.go`: Core model and state
- `models.go`: Component implementations
- `render.go`: View rendering
- `pages.go`: Page management
- `menus.go`: Menu definitions
- `dialog.go`: Dialog system
- Various controller implementations

## Reorganization Plan

The following directory structure is proposed to improve maintainability:

```
internal/cli/
├── components/  # UI components (table, form, theme, etc)
│   ├── form.go
│   ├── table.go
│   ├── theme.go
│   ├── navigation.go
│   ├── content.go
│   └── dialog.go
├── controllers/ # Page controllers and interfaces
│   ├── interfaces.go
│   ├── crud.go
│   └── specialized_controllers.go
├── pages/       # Page definitions and logic
│   ├── page_index.go
│   ├── page_registry.go
│   ├── data_pages.go
│   └── system_pages.go
├── styles/      # Theme and styling
│   ├── colors.go
│   ├── layout.go
│   └── titles.go
├── utils/       # Helper functions
│   ├── cursor.go
│   ├── validation.go
│   └── effects.go
├── core/        # Core model and update functionality
│   ├── model.go
│   ├── update.go
│   └── render.go
└── cli.go       # Package entry point and initialization
```

Implementation steps:
1. Create the directory structure
2. Move and refactor existing files to appropriate locations
3. Update imports throughout the package
4. Split large modules into smaller, focused components
5. Introduce consistent interfaces across components
6. Add tests for core functionality
7. Update documentation to reflect new organization