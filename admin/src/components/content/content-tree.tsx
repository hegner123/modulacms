import type { ContentNode } from '@modulacms/admin-sdk'
import { ScrollArea } from '@/components/ui/scroll-area'
import { TreeNode } from '@/components/content/tree-node'

type ContentTreeProps = {
  nodes: ContentNode[]
  selectedId: string | null
  onSelect: (id: string) => void
}

export function ContentTree({ nodes, selectedId, onSelect }: ContentTreeProps) {
  return (
    <ScrollArea className="h-full">
      <div className="p-2">
        {nodes.map((node) => (
          <TreeNode
            key={node.datatype.content.content_data_id}
            node={node}
            depth={0}
            selectedId={selectedId}
            onSelect={onSelect}
          />
        ))}
      </div>
    </ScrollArea>
  )
}
