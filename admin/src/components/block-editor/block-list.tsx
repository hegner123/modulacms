import { useMemo } from 'react'
import type { ContentNode, Datatype, ContentID } from '@modulacms/admin-sdk'
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
      dirty={dirty}
      saving={state.saving}
      deleting={deleteContent.isPending}
      depth={depth}
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
  const sortableIds = useMemo(
    () => childNodes.map((n) => n.datatype.content.content_data_id as string),
    [childNodes],
  )

  const { setNodeRef, isOver } = useDroppable({
    id: 'droppable-' + parentId,
    data: { parentId, type: 'container' },
  })

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'rounded-lg transition-colors',
        isOver && 'bg-primary/5 ring-2 ring-primary/20',
      )}
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
    </div>
  )
}
