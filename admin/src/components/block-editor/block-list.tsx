import { useMemo, useCallback } from 'react'
import type { ContentNode, Datatype, ContentID, ContentStatus, DatatypeID, UserID } from '@modulacms/admin-sdk'
import {
  DndContext,
  closestCenter,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  verticalListSortingStrategy,
  sortableKeyboardCoordinates,
} from '@dnd-kit/sortable'
import { useDeleteContentData, useUpdateContentData } from '@/queries/content'
import { useDatatypeFieldsByDatatype } from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { useAuthContext } from '@/lib/auth'
import { BlockCard } from './block-card'
import { BlockInserter } from './block-inserter'
import { buildMergedFields } from './build-merged-fields'
import type { useBlockEditorState } from './use-block-editor-state'

type BlockListProps = {
  childNodes: ContentNode[]
  datatypes: Datatype[]
  state: ReturnType<typeof useBlockEditorState>
  onInsert: (parentId: string, datatypeId: string) => void
  depth?: number
  parentId: string
}

function BlockCardWrapper({
  node,
  state,
  datatypes,
  onInsert,
  depth,
}: {
  node: ContentNode
  state: ReturnType<typeof useBlockEditorState>
  datatypes: Datatype[]
  onInsert: (parentId: string, datatypeId: string) => void
  depth: number
}) {
  const datatypeId = node.datatype.info.datatype_id
  const { data: datatypeFields } = useDatatypeFieldsByDatatype(datatypeId)
  const { data: allFields } = useFields()
  const deleteContent = useDeleteContentData()

  const mergedFields = useMemo(
    () => buildMergedFields(node, datatypeFields, allFields),
    [node, datatypeFields, allFields],
  )

  const contentDataId = node.datatype.content.content_data_id
  const selected = state.selectedBlockId === contentDataId
  const dirty = state.isBlockDirty(contentDataId, mergedFields)

  function renderNestedList(parentNode: ContentNode, childDepth: number) {
    const children = parentNode.nodes ?? []
    if (children.length === 0) return null
    return (
      <BlockList
        childNodes={children}
        datatypes={datatypes}
        state={state}
        onInsert={onInsert}
        depth={childDepth}
        parentId={parentNode.datatype.content.content_data_id}
      />
    )
  }

  return (
    <BlockCard
      node={node}
      mergedFields={mergedFields}
      selected={selected}
      dirty={dirty}
      saving={state.saving}
      deleting={deleteContent.isPending}
      depth={depth}
      onSelect={() => state.setSelectedBlockId(contentDataId)}
      onDeselect={() => state.setSelectedBlockId(null)}
      onSave={() => state.saveBlock(node, mergedFields)}
      onDelete={() => deleteContent.mutate(contentDataId as ContentID)}
      getFieldValue={state.getFieldValue}
      setFieldValue={state.setFieldValue}
      renderNestedList={renderNestedList}
    />
  )
}

export function BlockList({
  childNodes,
  datatypes,
  state,
  onInsert,
  depth = 0,
  parentId,
}: BlockListProps) {
  const { user } = useAuthContext()
  const updateContentData = useUpdateContentData()

  const sortableIds = useMemo(
    () => childNodes.map((n) => n.datatype.content.content_data_id as string),
    [childNodes],
  )

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 8 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event
      if (!over || active.id === over.id) return

      const oldIndex = sortableIds.indexOf(active.id as string)
      const newIndex = sortableIds.indexOf(over.id as string)
      if (oldIndex === -1 || newIndex === -1) return

      // Build the new order
      const newOrder = [...sortableIds]
      newOrder.splice(oldIndex, 1)
      newOrder.splice(newIndex, 0, active.id as string)

      // Find the moved node
      const movedNode = childNodes.find(
        (n) => n.datatype.content.content_data_id === active.id,
      )
      if (!movedNode) return

      // Compute new prev/next sibling IDs from new position
      const newPrev = newIndex > 0 ? newOrder[newIndex - 1] : null
      const newNext = newIndex < newOrder.length - 1 ? newOrder[newIndex + 1] : null

      const now = new Date().toISOString()
      updateContentData.mutate({
        content_data_id: active.id as ContentID,
        parent_id: movedNode.datatype.content.parent_id as ContentID | null,
        first_child_id: movedNode.datatype.content.first_child_id,
        next_sibling_id: newNext,
        prev_sibling_id: newPrev,
        route_id: (movedNode.datatype.content.route_id ?? null) as any,
        datatype_id: (movedNode.datatype.content.datatype_id ?? null) as DatatypeID | null,
        author_id: (user?.user_id ?? null) as UserID | null,
        status: movedNode.datatype.content.status as ContentStatus,
        date_created: movedNode.datatype.content.date_created,
        date_modified: now,
      })
    },
    [sortableIds, childNodes, updateContentData, user],
  )

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext items={sortableIds} strategy={verticalListSortingStrategy}>
        <div className="space-y-1">
          <BlockInserter datatypes={datatypes} onInsert={(dtId) => onInsert(parentId, dtId)} />
          {childNodes.map((child) => (
            <div key={child.datatype.content.content_data_id}>
              <BlockCardWrapper
                node={child}
                state={state}
                datatypes={datatypes}
                onInsert={onInsert}
                depth={depth}
              />
              <BlockInserter datatypes={datatypes} onInsert={(dtId) => onInsert(parentId, dtId)} />
            </div>
          ))}
        </div>
      </SortableContext>
    </DndContext>
  )
}
