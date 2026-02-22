import { useState, useMemo, useCallback } from 'react'
import type { ContentTree, ContentNode, ContentID, DatatypeID, UserID, ContentStatus, Route } from '@modulacms/admin-sdk'
import { TreePine } from 'lucide-react'
import {
  DndContext,
  DragOverlay,
  closestCenter,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core'
import { sortableKeyboardCoordinates } from '@dnd-kit/sortable'
import type { DatatypeField } from '@modulacms/admin-sdk'
import { useCreateContentData, useReorderContentData, useMoveContentData } from '@/queries/content'
import { useDatatypes, useDatatypeFieldsByDatatype, useDatatypeFields } from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { useAuthContext } from '@/lib/auth'
import { EmptyState } from '@/components/shared/empty-state'
import { BlockEditorHeader } from './block-editor-header'
import { PageFields } from './page-fields'
import { BlockList } from './block-list'
import { BlockInserter } from './block-inserter'
import { DocumentSettings } from './document-settings'
import { useBlockEditorState } from './use-block-editor-state'
import { buildMergedFields } from './build-merged-fields'

type BlockEditorProps = {
  route: Route
  treeData: ContentTree
}

function RootFieldsContainer({
  rootNode,
  state,
}: {
  rootNode: ContentNode
  state: ReturnType<typeof useBlockEditorState>
}) {
  const datatypeId = rootNode.datatype.info.datatype_id
  const { data: datatypeFields } = useDatatypeFieldsByDatatype(datatypeId)
  const { data: allFields } = useFields()

  const mergedFields = useMemo(
    () => buildMergedFields(rootNode, datatypeFields, allFields),
    [rootNode, datatypeFields, allFields],
  )

  return (
    <PageFields
      node={rootNode}
      mergedFields={mergedFields}
      getFieldValue={state.getFieldValue}
      setFieldValue={state.setFieldValue}
    />
  )
}

function countAllBlocks(node: ContentNode): number {
  const children = node.nodes ?? []
  let count = children.length
  for (const child of children) {
    count += countAllBlocks(child)
  }
  return count
}

/** Collect the ordered child IDs for a given parent from the tree. */
function findChildIds(tree: ContentNode, parentId: string): string[] {
  if (tree.datatype.content.content_data_id === parentId) {
    return (tree.nodes ?? []).map((n) => n.datatype.content.content_data_id)
  }
  for (const child of tree.nodes ?? []) {
    const result = findChildIds(child, parentId)
    if (result.length > 0) return result
  }
  return []
}

/** Recursively collect all nodes (including the given node) from the tree. */
function collectAllNodes(node: ContentNode): ContentNode[] {
  const result: ContentNode[] = [node]
  for (const child of node.nodes ?? []) {
    result.push(...collectAllNodes(child))
  }
  return result
}

/** Find the datatype label for a given content data ID in the tree. */
function findDatatypeLabel(node: ContentNode, contentDataId: string): string | null {
  if (node.datatype.content.content_data_id === contentDataId) {
    return node.datatype.info.label
  }
  for (const child of node.nodes ?? []) {
    const result = findDatatypeLabel(child, contentDataId)
    if (result) return result
  }
  return null
}

export function BlockEditor({ route, treeData }: BlockEditorProps) {
  const { user } = useAuthContext()
  const { data: datatypes } = useDatatypes()
  const { data: allFields } = useFields()
  const { data: allDatatypeFields } = useDatatypeFields()
  const createContent = useCreateContentData()
  const reorderContentData = useReorderContentData()
  const moveContentData = useMoveContentData()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [activeId, setActiveId] = useState<string | null>(null)
  const state = useBlockEditorState()

  const tree = treeData as ContentTree
  const rootNode = tree.root
  const childNodes = rootNode?.nodes ?? []

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 8 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }, [])

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      setActiveId(null)
      const { active, over } = event
      if (!over || active.id === over.id) return
      if (!rootNode) return

      const activeParentId = (active.data.current as { parentId?: string })?.parentId
      const overData = over.data.current as { parentId?: string; type?: string }
      let overParentId = overData?.parentId

      // If dropped on a droppable container, use that container's parentId
      if (overData?.type === 'container') {
        overParentId = overData.parentId
      }

      if (activeParentId === overParentId) {
        // Same parent — reorder
        const siblingIds = findChildIds(rootNode, activeParentId ?? '')
        const oldIndex = siblingIds.indexOf(active.id as string)
        const newIndex = siblingIds.indexOf(over.id as string)
        if (oldIndex === -1 || newIndex === -1) return

        const newOrder = [...siblingIds]
        newOrder.splice(oldIndex, 1)
        newOrder.splice(newIndex, 0, active.id as string)

        reorderContentData.mutate({
          parent_id: (activeParentId || null) as ContentID | null,
          ordered_ids: newOrder as ContentID[],
        })
      } else {
        // Different parent — move
        const destParentId = overParentId ?? ''
        const destChildIds = findChildIds(rootNode, destParentId)

        // Compute position: place after the over item, or at position 0 for containers
        let position = 0
        if (overData?.type === 'container') {
          position = destChildIds.length
        } else {
          const overIndex = destChildIds.indexOf(over.id as string)
          position = overIndex === -1 ? destChildIds.length : overIndex
        }

        moveContentData.mutate({
          node_id: active.id as ContentID,
          new_parent_id: (destParentId || null) as ContentID | null,
          position,
        })
      }
    },
    [rootNode, reorderContentData, moveContentData],
  )

  if (!rootNode) {
    return (
      <EmptyState
        icon={TreePine}
        title="No content"
        description="This route has no content data yet."
      />
    )
  }

  function handleInsert(parentId: string, datatypeId: string) {
    const now = new Date().toISOString()
    createContent.mutate({
      route_id: route.route_id,
      parent_id: parentId as unknown as ContentID,
      first_child_id: null,
      next_sibling_id: null,
      prev_sibling_id: null,
      datatype_id: datatypeId as DatatypeID,
      author_id: (user?.user_id ?? null) as UserID | null,
      status: 'draft' as ContentStatus,
      date_created: now,
      date_modified: now,
    })
  }

  const rootContentDataId = rootNode.datatype.content.content_data_id
  const activeDatatypeLabel = activeId ? findDatatypeLabel(rootNode, activeId) : null

  // Build a per-datatype lookup for datatype fields
  const datatypeFieldMap = useMemo(() => {
    const map = new Map<string, DatatypeField[]>()
    if (!allDatatypeFields) return map
    for (const df of allDatatypeFields) {
      const list = map.get(df.datatype_id) ?? []
      list.push(df)
      map.set(df.datatype_id, list)
    }
    return map
  }, [allDatatypeFields])

  function handleSaveAll() {
    const allNodes = collectAllNodes(rootNode)
    const blocks = allNodes.map((node) => {
      const dtId = node.datatype.info.datatype_id
      const dtFields = datatypeFieldMap.get(dtId)
      return {
        node,
        mergedFields: buildMergedFields(node, dtFields, allFields),
      }
    })
    state.saveAllForce(blocks)
  }

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col">
      <BlockEditorHeader
        title={route.title}
        slug={route.slug}
        dirtyCount={0}
        saving={state.saving}
        settingsOpen={settingsOpen}
        onToggleSettings={() => setSettingsOpen(!settingsOpen)}
        onSaveAll={handleSaveAll}
      />

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-auto">
          <div className="mx-auto max-w-3xl space-y-6 px-6 py-8">
            <RootFieldsContainer rootNode={rootNode} state={state} />

            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
            >
              {childNodes.length > 0 ? (
                <BlockList
                  childNodes={childNodes}
                  datatypes={datatypes ?? []}
                  state={state}
                  onInsert={handleInsert}
                  parentId={rootContentDataId}
                />
              ) : (
                <div className="py-8 text-center">
                  <p className="mb-4 text-sm text-muted-foreground">
                    No content blocks yet. Use the + button to add one.
                  </p>
                  <BlockInserter
                    datatypes={datatypes ?? []}
                    onInsert={(dtId) => handleInsert(rootContentDataId, dtId)}
                  />
                </div>
              )}
              <DragOverlay>
                {activeId && activeDatatypeLabel ? (
                  <div className="rounded-md border border-primary/50 bg-primary/10 px-3 py-1.5 text-xs font-medium text-primary shadow-sm">
                    {activeDatatypeLabel}
                  </div>
                ) : null}
              </DragOverlay>
            </DndContext>
          </div>
        </div>

        {settingsOpen && (
          <DocumentSettings
            rootNode={rootNode}
            blockCount={countAllBlocks(rootNode)}
          />
        )}
      </div>
    </div>
  )
}
