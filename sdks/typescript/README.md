# ModulaCMS TypeScript SDKs

> A Go backend walks into a bar and orders four TypeScript packages. The bartender says "we don't serve your type here." The backend says "that's fine, I brought my own."

## The Family

This is a pnpm monorepo containing everything you need to talk to a ModulaCMS instance from TypeScript — or just rearrange trees for fun.

| Package | npm | What it does | Vibe |
|---------|-----|-------------|------|
| [`@modulacms/types`](./types/) | [![npm](https://img.shields.io/npm/v/@modulacms/types)](https://www.npmjs.com/package/@modulacms/types) | Shared types, branded IDs, enums | The boring one that everyone depends on but nobody thanks |
| [`@modulacms/sdk`](./modulacms-sdk/) | [![npm](https://img.shields.io/npm/v/@modulacms/sdk)](https://www.npmjs.com/package/@modulacms/sdk) | Read-only content delivery | Look but don't touch |
| [`@modulacms/admin-sdk`](./modulacms-admin-sdk/) | [![npm](https://img.shields.io/npm/v/@modulacms/admin-sdk)](https://www.npmjs.com/package/@modulacms/admin-sdk) | Full admin CRUD | Touch everything, break nothing (hopefully) |
| [`@modulacms/tree`](./tree/) | [![npm](https://img.shields.io/npm/v/@modulacms/tree)](https://www.npmjs.com/package/@modulacms/tree) | Sibling-pointer tree ops | For people who think linked lists aren't confusing enough |

## Quick Start

```bash
pnpm install
pnpm build        # types first, then the rest in parallel — just like a healthy family
pnpm test         # trust but verify
pnpm typecheck    # because "it works on my machine" isn't a type
```

## Architecture

```
You are here
    |
    v
sdks/typescript/
    |
    |-- types/              <-- The foundation. 0 dependencies. Pure vibes.
    |     |
    |     +-- Branded IDs, enums, entity types
    |
    |-- tree/               <-- Also 0 dependencies. Doesn't even know types/ exists.
    |     |                     They're like coworkers who've never met.
    |     +-- Sibling-pointer tree mutations & queries
    |
    |-- modulacms-sdk/      <-- Depends on types/. Reads content. Very polite.
    |     |
    |     +-- GET requests and a dream
    |
    |-- modulacms-admin-sdk/ <-- Depends on types/. Full CRUD. Has admin energy.
          |
          +-- 150+ methods and not a single regret
```

### The Build Order Situation

`types` and `tree` build first (in parallel — they're independent spirits). Then `sdk` and `admin-sdk` build in parallel because they both depend on `types` but not on each other. pnpm handles this automatically. You don't need to think about it. In fact, please don't think about it. It's handled.

## The Tree Package Deserves Its Own Section

`@modulacms/tree` implements a sibling-pointer tree structure where every node has `parentId`, `firstChildId`, `nextSiblingId`, and `prevSiblingId`. If that sounds like a linked list that made questionable life choices — you're not wrong, but it enables O(1) insertions and reorderings, which is more than you can say for most of your data structures.

```typescript
import { createTree, createNode, addNode, appendChild, getChildren } from '@modulacms/tree'

const tree = createTree()
const root = createNode('root')
const child = createNode('child')

addNode(tree, root)
addNode(tree, child)
tree.rootId = root.id
appendChild(tree, child.id, root.id)

getChildren(tree, root.id) // [child] — parenting is easy actually
```

Zero dependencies. Generic over any node shape. Mutates in place because immutability is a spectrum and this library chose violence.

## Development

```bash
just build          # Build all packages
just test           # Run all tests
just typecheck      # Strict mode. noUncheckedIndexedAccess. We don't play.
just clean          # Start fresh, like a Monday

just patch all      # Bump all patch versions
just release all    # Patch + build + publish. YOLO responsibly.
```

### Adding a Package

1. Create a directory
2. Add it to `pnpm-workspace.yaml`
3. Give it a `package.json`, `tsconfig.json`, and `tsup.config.ts`
4. Run `pnpm install`
5. Contemplate whether the JavaScript ecosystem really needed another package
6. Ship it anyway

## FAQ

**Q: Why are there SDKs in TypeScript, Go, AND Swift?**
A: Because the backend is written in Go, the admin panel uses HTMX, the block editor is vanilla JS, and at some point someone said "what if phones." We contain multitudes.

**Q: Why does `@modulacms/tree` not depend on `@modulacms/types`?**
A: Because not every tree needs to know it's carrying CMS content. Sometimes a tree just wants to be a tree. Let it be a tree.

**Q: Is this production-ready?**
A: The types are. The SDKs are. The tree just got here but it brought 76 tests, which is more preparation than most of us bring to a Monday standup.

**Q: Why pnpm?**
A: `node_modules` was 800MB with npm. With pnpm it's a symlink farm that takes up approximately nothing. We like nothing.

## License

MIT. Do whatever you want. We're a CMS, not a lawyer.
