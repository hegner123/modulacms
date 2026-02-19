import { useState } from 'react'
import type { DraggableAttributes } from '@dnd-kit/core'
import type { SyntheticListenerMap } from '@dnd-kit/core/dist/hooks/utilities'
import { ChevronRight, GripVertical, Trash2, Save } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { cn } from '@/lib/utils'

type BlockToolbarProps = {
  datatypeLabel: string
  dirty: boolean
  saving: boolean
  deleting: boolean
  childCount: number
  expanded: boolean
  dragHandleListeners: SyntheticListenerMap | undefined
  dragHandleAttributes: DraggableAttributes
  setActivatorNodeRef: (element: HTMLElement | null) => void
  onToggleExpand: () => void
  onSave: () => void
  onDelete: () => void
}

export function BlockToolbar({
  datatypeLabel,
  dirty,
  saving,
  deleting,
  childCount,
  expanded,
  dragHandleListeners,
  dragHandleAttributes,
  setActivatorNodeRef,
  onToggleExpand,
  onSave,
  onDelete,
}: BlockToolbarProps) {
  const [confirmOpen, setConfirmOpen] = useState(false)

  return (
    <>
      <div className="flex items-center gap-2 px-4 py-2">
        <button
          ref={setActivatorNodeRef}
          className="cursor-grab touch-none rounded p-0.5 text-muted-foreground hover:text-foreground active:cursor-grabbing"
          {...dragHandleListeners}
          {...dragHandleAttributes}
        >
          <GripVertical className="h-4 w-4" />
        </button>
        {childCount > 0 && (
          <Button
            size="icon"
            variant="ghost"
            className="h-6 w-6"
            onClick={(e) => {
              e.stopPropagation()
              onToggleExpand()
            }}
          >
            <ChevronRight className={cn(
              'h-3.5 w-3.5 transition-transform',
              expanded && 'rotate-90',
            )} />
          </Button>
        )}
        <Badge variant="outline" className="text-xs">
          {datatypeLabel}
        </Badge>
        {childCount > 0 && (
          <span className="text-xs text-muted-foreground">
            {childCount} {childCount === 1 ? 'child' : 'children'}
          </span>
        )}
        <div className="flex-1" />
        {dirty && (
          <Button size="sm" variant="ghost" onClick={onSave} disabled={saving}>
            <Save className="mr-1 h-3 w-3" />
            {saving ? 'Saving...' : 'Save'}
          </Button>
        )}
        <Button
          size="icon"
          variant="ghost"
          className="h-7 w-7 text-destructive hover:text-destructive"
          onClick={() => setConfirmOpen(true)}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Delete Block"
        description="This will permanently delete this content block and all its fields. This action cannot be undone."
        onConfirm={onDelete}
        loading={deleting}
      />
    </>
  )
}
