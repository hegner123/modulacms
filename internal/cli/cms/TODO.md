# CMS Implementation TODO List

## Database Operations
- [ ] Implement datatype definition creation
- [ ] Implement content data instance creation
- [ ] Implement datatype definition reading
- [ ] Implement content data instance reading
- [ ] Implement datatype definition updating
  - [ ] Add CRUD operations for fields
- [ ] Implement content data instance updating
  - [ ] Add CRUD operations for fields
- [ ] Implement datatype definition deletion
- [ ] Implement content data instance deletion

## UI Components
- [ ] Implement page preview component
- [ ] Implement dialog system for CMS operations

## Keyboard Controls
- [ ] Add key binding for adding datatypes (A)
- [ ] Add key binding for deleting datatypes (X)
- [ ] Add key binding for redo (Y)
- [ ] Add key binding for undo (Z)
- [ ] Add key binding for save (S)
- [ ] Add key binding for showing fields of selected datatype (F)
- [ ] Add key binding for editing highlighted item (I/E)
- [ ] Add key binding for duplicating selected datatype (D)
- [ ] Add key binding for copying datatype/field (C)
- [ ] Add key binding for pasting datatype/field (P)
- [ ] Add key binding for navigating up tree (J)
- [ ] Add key binding for navigating down tree (K)
- [ ] Add key binding for navigating up node children (H)
- [ ] Add key binding for navigating down node children (L)

## Dialog Components
- [ ] Create datatype selection dialog
- [ ] Create datatype options dialog
- [ ] Create field selection dialog
- [ ] Create field input dialogs:
  - [ ] Number input
  - [ ] Text input
  - [ ] Validated text input
  - [ ] Text area input
  - [ ] Media picker

## Content Model Implementation
- [ ] Create MODEL struct for assembling content
- [ ] Implement content data row creation as Root of specified datatype
- [ ] Create functionality to write MODEL struct to database (content data and content field rows)
- [ ] Implement functionality to list datatypes approved as children of parent datatype
- [ ] Implement functionality to list fields of datatype as menu
- [ ] Add cursor-based field selection and input
- [ ] Implement database save after field input

## Integration with Existing CLI System
- [ ] Connect CMS functionality with existing CLI page system
- [ ] Add CMS pages to navigation
- [ ] Ensure proper state management between CMS and main CLI interface

## Testing
- [ ] Write tests for datatype CRUD operations
- [ ] Write tests for content data CRUD operations
- [ ] Test UI interactions and keyboard controls
- [ ] Test dialog components
- [ ] Test content model creation and storage
