# Dynamic Menu Implementation Plan

## Phase 1: Message System Refactor
1. **Create proper async message types** in `actions.go`:
   - `fetchDatatypesCmd` - returns `tea.Cmd` 
   - `datatypesFetchedMsg` - carries fetched datatypes
   - `datatypesFetchErrorMsg` - carries errors

2. **Fix database query pattern** in `actions.go`:
   - Move `CLIRead` logic into proper `tea.Cmd` function
   - Return message instead of blocking execution
   - Add proper error handling

## Phase 2: Menu State Management
3. **Add menu state tracking** to Model:
   - `DynamicMenus map[string][]*Page` - track menus by context
   - `MenuState` field to track current dynamic menu state
   - Clear methods for menu cleanup

4. **Update menu building** in `menus.go`:
   - `BuildDatatypeMenu` should replace, not append
   - Add menu cleanup functions
   - Implement menu caching to avoid repeated DB calls

## Phase 3: Controller Integration
5. **Implement `datatypeInterface` handler** in `update.go`:
   - Add `DatatypeControls` method
   - Handle datatype-specific navigation
   - Integrate with existing control patterns

6. **Fix page hierarchy** in `pages.go`:
   - Set proper parent relationships for dynamic pages
   - Update page navigation to handle dynamic children
   - Ensure cleanup when leaving dynamic sections

## Phase 4: Router Updates
7. **Refactor router logic** in `router.go`:
   - Move database fetching to proper message handlers
   - Clean separation between routing and data fetching
   - Handle dynamic page navigation properly

## Testing Strategy
- Test menu state transitions
- Verify no memory leaks from dynamic pages
- Ensure proper error handling for DB failures
- Test navigation between static/dynamic menus

## Key Troublespots to Address
- Blocking database queries in Update function
- Menu state accumulation without cleanup
- Missing controller implementation for dynamic pages
- Improper message flow patterns
- Page hierarchy confusion for dynamic items
