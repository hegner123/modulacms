import { useState } from 'react'
import type { ContentNode, ContentStatus, Datatype } from '@modulacms/admin-sdk'
import { useSortable } from '@dnd-kit/sortable'
import { cn } from '@/lib/utils'
import { BlockToolbar } from './block-toolbar'
import { BlockFieldEditor } from './block-field-editor'
import { BlockInserter } from './block-inserter'
import type { MergedField } from './build-merged-fields'

type BlockCardProps = {
  node: ContentNode
  mergedFields: MergedField[]
  dirty: boolean
  saving: boolean
  deleting: boolean
  statusUpdating: boolean
  depth: number
  datatypes: Datatype[]
  onSave: () => void
  onDelete: () => void
  onStatusChange: (status: ContentStatus) => void
  onInsertChild: (parentId: string, datatypeId: string, prevSiblingId?: string | null, nextSiblingId?: string | null) => void
  getFieldValue: (contentDataId: string, fieldId: string, mergedFields: MergedField[]) => string
  setFieldValue: (contentDataId: string, fieldId: string, value: string) => void
  renderNestedList: (parentNode: ContentNode, depth: number) => React.ReactNode
}

export function BlockCard({
  node,
  mergedFields,
  dirty,
  saving,
  deleting,
  statusUpdating,
  depth,
  datatypes,
  onSave,
  onDelete,
  onStatusChange,
  onInsertChild,
  getFieldValue,
  setFieldValue,
  renderNestedList,
}: BlockCardProps) {
  const contentDataId = node.datatype.content.content_data_id
  const childNodes = node.nodes ?? []
  const [expanded, setExpanded] = useState(childNodes.length > 0)

  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    isDragging,
  } = useSortable({
    id: contentDataId,
    data: { parentId: node.datatype.content.parent_id, type: 'item' },
  })

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'rounded-lg border border-border',
        isDragging && 'opacity-25',
      )}
    >
      <BlockToolbar
        datatypeLabel={node.datatype.info.label}
        status={node.datatype.content.status}
        dirty={dirty}
        saving={saving}
        deleting={deleting}
        statusUpdating={statusUpdating}
        childCount={childNodes.length}
        expanded={expanded}
        dragHandleListeners={listeners}
        dragHandleAttributes={attributes}
        setActivatorNodeRef={setActivatorNodeRef}
        onToggleExpand={() => setExpanded(!expanded)}
        onSave={onSave}
        onDelete={onDelete}
        onStatusChange={onStatusChange}
      />

      <BlockFieldEditor
        contentDataId={contentDataId}
        mergedFields={mergedFields}
        getFieldValue={getFieldValue}
        setFieldValue={setFieldValue}
      />

      {expanded && (
        <div className="border-t pl-6">
          {childNodes.length > 0 ? (
            renderNestedList(node, depth + 1)
          ) : (
            <div className="py-2">
              <BlockInserter
                datatypes={datatypes}
                onInsert={(dtId) => onInsertChild(contentDataId, dtId)}
              />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
