import { useState } from 'react'
import type { ContentNode, ContentStatus } from '@modulacms/admin-sdk'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { cn } from '@/lib/utils'
import { BlockToolbar } from './block-toolbar'
import { BlockFieldEditor } from './block-field-editor'
import type { MergedField } from './build-merged-fields'

type BlockCardProps = {
  node: ContentNode
  mergedFields: MergedField[]
  dirty: boolean
  saving: boolean
  deleting: boolean
  statusUpdating: boolean
  depth: number
  onSave: () => void
  onDelete: () => void
  onStatusChange: (status: ContentStatus) => void
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
  onSave,
  onDelete,
  onStatusChange,
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
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: contentDataId,
    data: { parentId: node.datatype.content.parent_id, type: 'item' },
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'rounded-lg border border-border',
        isDragging && 'z-50 opacity-50 shadow-lg',
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

      {expanded && childNodes.length > 0 && (
        <div className="border-t pl-6">
          {renderNestedList(node, depth + 1)}
        </div>
      )}
    </div>
  )
}
