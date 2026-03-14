# TUI Media Field Spec

## Current State

Media fields (`type: media`) render as a plain `TextInputBubble` where users type a ULID manually. This is technically functional but unusable — no one knows media ULIDs by heart.

## Desired Behavior

When a content form has a `media` type field, render two buttons instead of a text input:

```
Featured Image
┌──────────────────────────────────┐
│  [Browse]  [Upload]              │
│  hero-banner.jpg (image/jpeg)    │
└──────────────────────────────────┘
```

### Browse Button
- Opens the Media Library picker (not the filesystem file picker)
- Shows existing media items with **name** and **file type** (mimetype)
- User selects an item → field value set to the media ULID
- Display shows the selected media's name + type below the buttons

### Upload Button
- Opens the file picker (existing `filepicker` component)
- Uploads the selected file via the media upload pipeline
- After upload completes, field value set to the new media ULID
- Display updates with the uploaded file's name + type

### Display When Set
- Show media name (e.g., `hero-banner.jpg`)
- Show mimetype (e.g., `image/jpeg`)
- No image preview (terminal limitation)
- Show a `[Clear]` action to remove the selection

### Display When Empty
- Show `[Browse]  [Upload]` buttons
- Show `(no media selected)` placeholder

## Implementation Notes

### New Bubble Type: `MediaBubble`

Create `bubble_media.go` implementing `FieldBubble`:
- State: selected media ID, media name, media mimetype
- Render: buttons + selected file info
- Keys:
  - `enter` on Browse → emit `MediaBrowseRequestMsg`
  - `enter` on Upload → emit file picker open
  - `tab` cycles between Browse/Upload/Clear
  - `backspace` or `delete` on Clear → clear selection

### Media Library Picker

New overlay (or reuse existing media tree from MediaScreen):
- List media items showing: name, mimetype, file size
- Up/down navigation, enter to select, escape to cancel
- Returns `MediaSelectedMsg{MediaID, Name, Mimetype}`

### Field Type Registry

Register `media` type with `MediaBubble` instead of `TextInputBubble`:
```go
// In type_registry.go init()
Register("media", "Media", func() FieldBubble { return &MediaBubble{} })
```

### IDRef Fields

Same pattern applies to `idref` type fields — they reference other content items. Could use a similar picker showing content titles instead of ULIDs. Lower priority than media.

## Scope

- Phase 1: MediaBubble with Browse (media library picker) + display name/type
- Phase 2: Upload button (file picker + upload pipeline integration)
- Phase 3: IDRef picker for content references
