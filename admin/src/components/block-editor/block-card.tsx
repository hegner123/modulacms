import { useRef, useEffect, useState, useCallback } from 'react'
import type { ContentNode } from '@modulacms/admin-sdk'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { cn } from '@/lib/utils'
import { BlockToolbar } from './block-toolbar'
import { BlockFieldEditor } from './block-field-editor'
import type { MergedField } from './build-merged-fields'

type BlockCardProps = {
  node: ContentNode
  mergedFields: MergedField[]
  selected: boolean
  dirty: boolean
  saving: boolean
  deleting: boolean
  depth: number
  onSelect: () => void
  onDeselect: () => void
  onSave: () => void
  onDelete: () => void
  getFieldValue: (contentDataId: string, fieldId: string, mergedFields: MergedField[]) => string
  setFieldValue: (contentDataId: string, fieldId: string, value: string) => void
  renderNestedList: (parentNode: ContentNode, depth: number) => React.ReactNode
}

function getPreviewText(mergedFields: MergedField[], getFieldValue: BlockCardProps['getFieldValue'], contentDataId: string): string {
  const TEXT_TYPES = new Set(['text', 'slug', 'textarea', 'richtext', 'email', 'url'])
  for (const field of mergedFields) {
    if (TEXT_TYPES.has(field.type)) {
      const val = getFieldValue(contentDataId, field.fieldId, mergedFields)
      if (val) return val.length > 120 ? val.slice(0, 120) + '...' : val
    }
  }
  return ''
}

export function BlockCard({
  node,
  mergedFields,
  selected,
  dirty,
  saving,
  deleting,
  depth,
  onSelect,
  onDeselect,
  onSave,
  onDelete,
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
  } = useSortable({ id: contentDataId })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  // Click-outside to deselect — use a separate ref merged with sortable ref
  const clickOutsideRef = useRef<HTMLDivElement | null>(null)
  const mergedRef = useCallback(
    (el: HTMLDivElement | null) => {
      clickOutsideRef.current = el
      setNodeRef(el)
    },
    [setNodeRef],
  )

  // Stable ref so the effect doesn't re-register on every render
  const onDeselectRef = useRef(onDeselect)
  onDeselectRef.current = onDeselect

  useEffect(() => {
    if (!selected) return

    function handleClickOutside(e: MouseEvent) {
      const target = e.target as Element
      // Ignore clicks inside portals (dialogs, popovers, etc.)
      if (target.closest?.('[data-radix-portal]')) return
      if (clickOutsideRef.current && !clickOutsideRef.current.contains(target)) {
        onDeselectRef.current()
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [selected])

  return (
    <div
      ref={mergedRef}
      style={style}
      className={cn(
        'group/card rounded-lg border transition-colors',
        selected
          ? 'border-primary ring-1 ring-primary'
          : 'border-transparent hover:border-border',
        isDragging && 'z-50 opacity-50 shadow-lg',
      )}
      onClick={(e) => {
        if (!selected) {
          e.stopPropagation()
          onSelect()
        }
      }}
    >
      <div className={cn(
        'transition-opacity',
        selected
          ? 'opacity-100'
          : 'pointer-events-none opacity-0 group-hover/card:pointer-events-auto group-hover/card:opacity-100',
      )}>
        <BlockToolbar
          datatypeLabel={node.datatype.info.label}
          dirty={dirty}
          saving={saving}
          deleting={deleting}
          childCount={childNodes.length}
          expanded={expanded}
          dragHandleListeners={listeners}
          dragHandleAttributes={attributes}
          setActivatorNodeRef={setActivatorNodeRef}
          onToggleExpand={() => setExpanded(!expanded)}
          onSave={onSave}
          onDelete={onDelete}
        />
      </div>

      {selected ? (
        <BlockFieldEditor
          contentDataId={contentDataId}
          mergedFields={mergedFields}
          getFieldValue={getFieldValue}
          setFieldValue={setFieldValue}
        />
      ) : (() => {
        const preview = getPreviewText(mergedFields, getFieldValue, contentDataId)
        return preview ? (
          <div className="px-4 pb-3">
            <p className="text-sm text-muted-foreground">{preview}</p>
          </div>
        ) : null
      })()}

      {expanded && childNodes.length > 0 && (
        <div className="border-t pl-6">
          {renderNestedList(node, depth + 1)}
        </div>
      )}
    </div>
  )
}
