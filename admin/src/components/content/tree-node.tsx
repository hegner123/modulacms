import { useState } from 'react'
import type { ContentNode } from '@modulacms/admin-sdk'
import { ChevronRight, ChevronDown, FileText, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'

type TreeNodeProps = {
  node: ContentNode
  depth: number
  selectedId: string | null
  onSelect: (id: string) => void
}

export function TreeNode({ node, depth, selectedId, onSelect }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(true)
  const children = node.nodes ?? []
  const hasChildren = children.length > 0
  const isSelected = selectedId === node.datatype.content.content_data_id

  function handleToggle(e: React.MouseEvent) {
    e.stopPropagation()
    setExpanded((prev) => !prev)
  }

  function handleSelect() {
    onSelect(node.datatype.content.content_data_id)
  }

  return (
    <div>
      <button
        type="button"
        onClick={handleSelect}
        className={cn(
          'flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground',
          isSelected && 'bg-primary text-primary-foreground',
        )}
        style={{ paddingLeft: `${depth * 20 + 8}px` }}
      >
        {hasChildren ? (
          <span
            role="button"
            tabIndex={-1}
            onClick={handleToggle}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.stopPropagation()
                setExpanded((prev) => !prev)
              }
            }}
            className="flex shrink-0 items-center justify-center rounded p-0.5 hover:bg-muted"
          >
            {expanded ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
          </span>
        ) : (
          <span className="w-5" />
        )}
        {hasChildren ? (
          <FolderOpen className="h-4 w-4 shrink-0 text-muted-foreground" />
        ) : (
          <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
        )}
        <span className="truncate">{node.datatype.info.label}</span>
      </button>
      {hasChildren && expanded && (
        <div>
          {children.map((child) => (
            <TreeNode
              key={child.datatype.content.content_data_id}
              node={child}
              depth={depth + 1}
              selectedId={selectedId}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  )
}
