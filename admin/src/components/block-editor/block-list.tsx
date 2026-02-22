import { useMemo } from 'react'
import type { ContentNode, ContentStatus, Datatype, ContentID } from '@modulacms/admin-sdk'
import { useDroppable } from '@dnd-kit/core'
import {
  SortableContext,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { useDeleteContentData } from '@/queries/content'
import { useDatatypeFieldsByDatatype } from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { cn } from '@/lib/utils'
import { BlockCard } from './block-card'
import { BlockInserter } from './block-inserter'
import { buildMergedFields } from './build-merged-fields'
import type { useBlockEditorState } from './use-block-editor-state'

type InsertFn = (parentId: string, datatypeId: string, prevSiblingId?: string | null, nextSiblingId?: string | null) => void

type BlockListProps = {
  childNodes: ContentNode[]
  datatypes: Datatype[]
  state: ReturnType<typeof useBlockEditorState>
  onInsert: InsertFn
  onStatusChange: (node: ContentNode, status: ContentStatus) => void
  statusUpdatingIds: Set<string>
  depth?: number
  parentId: string
  activeId?: string | null
  overId?: string | null
}

function BlockCardWrapper({
  node,
  state,
  datatypes,
  onInsert,
  onStatusChange,
  statusUpdatingIds,
  depth,
  activeId,
  overId,
}: {
  node: ContentNode
  state: ReturnType<typeof useBlockEditorState>
  datatypes: Datatype[]
  onInsert: InsertFn
  onStatusChange: (node: ContentNode, status: ContentStatus) => void
  statusUpdatingIds: Set<string>
  depth: number
  activeId?: string | null
  overId?: string | null
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
        onStatusChange={onStatusChange}
        statusUpdatingIds={statusUpdatingIds}
        depth={childDepth}
        parentId={parentNode.datatype.content.content_data_id}
        activeId={activeId}
        overId={overId}
      />
    )
  }

  return (
    <BlockCard
      node={node}
      mergedFields={mergedFields}
      dirty={dirty}
      saving={state.saving}
      deleting={deleteContent.isPending}
      statusUpdating={statusUpdatingIds.has(contentDataId)}
      depth={depth}
      datatypes={datatypes}
      onSave={() => state.saveBlock(node, mergedFields)}
      onDelete={() => deleteContent.mutate(contentDataId as ContentID)}
      onStatusChange={(status) => onStatusChange(node, status)}
      onInsertChild={onInsert}
      getFieldValue={state.getFieldValue}
      setFieldValue={state.setFieldValue}
      renderNestedList={renderNestedList}
    />
  )
}

function DropIndicator() {
  return (
    <div className="flex items-center gap-2 py-0.5">
      <div className="h-0.5 flex-1 rounded-full bg-primary" />
    </div>
  )
}

export function BlockList({
  childNodes,
  datatypes,
  state,
  onInsert,
  onStatusChange,
  statusUpdatingIds,
  depth = 0,
  parentId,
  activeId,
  overId,
}: BlockListProps) {
  const sortableIds = useMemo(
    () => childNodes.map((n) => n.datatype.content.content_data_id as string),
    [childNodes],
  )

  const droppableId = 'droppable-' + parentId
  const { setNodeRef, isOver } = useDroppable({
    id: droppableId,
    data: { parentId, type: 'container' },
  })

  // Show the container highlight when something is dragged over this container's
  // droppable zone (not over a child item within it).
  const isDragging = activeId !== null && activeId !== undefined
  const containerHighlight = isDragging && isOver && overId === droppableId

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'rounded-lg transition-colors',
        containerHighlight && 'ring-2 ring-primary/30',
      )}
    >
      <SortableContext items={sortableIds} strategy={verticalListSortingStrategy}>
        <div className="space-y-1">
          <BlockInserter
            datatypes={datatypes}
            onInsert={(dtId) => onInsert(
              parentId, dtId, null,
              childNodes[0]?.datatype.content.content_data_id ?? null,
            )}
          />
          {childNodes.map((child, i) => {
            const childId = child.datatype.content.content_data_id
            const nextChild = childNodes[i + 1] ?? null
            const showIndicator = isDragging && overId === childId && activeId !== childId
            return (
              <div key={childId}>
                {showIndicator && <DropIndicator />}
                <BlockCardWrapper
                  node={child}
                  state={state}
                  datatypes={datatypes}
                  onInsert={onInsert}
                  onStatusChange={onStatusChange}
                  statusUpdatingIds={statusUpdatingIds}
                  depth={depth}
                  activeId={activeId}
                  overId={overId}
                />
                <BlockInserter
                  datatypes={datatypes}
                  onInsert={(dtId) => onInsert(
                    parentId, dtId, childId,
                    nextChild?.datatype.content.content_data_id ?? null,
                  )}
                />
              </div>
            )
          })}
          {containerHighlight && <DropIndicator />}
        </div>
      </SortableContext>
    </div>
  )
}
