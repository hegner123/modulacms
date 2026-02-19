import type { ContentNode } from '@modulacms/admin-sdk'
import { Badge } from '@/components/ui/badge'
import { formatDateTime } from '@/lib/utils'

type DocumentSettingsProps = {
  rootNode: ContentNode
  blockCount: number
}

export function DocumentSettings({ rootNode, blockCount }: DocumentSettingsProps) {
  const content = rootNode.datatype.content

  return (
    <div className="w-80 shrink-0 border-l">
      <div className="p-4">
        <h3 className="mb-4 text-sm font-semibold">Document Settings</h3>

        <div className="space-y-4">
          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">Status</p>
            <Badge variant={content.status === 'published' ? 'default' : 'secondary'}>
              {content.status}
            </Badge>
          </div>

          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">Created</p>
            <p className="text-sm">{formatDateTime(content.date_created)}</p>
          </div>

          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">Modified</p>
            <p className="text-sm">{formatDateTime(content.date_modified)}</p>
          </div>

          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">Blocks</p>
            <p className="text-sm">{blockCount}</p>
          </div>

          <div className="space-y-1">
            <p className="text-xs font-medium text-muted-foreground">Content ID</p>
            <p className="break-all font-mono text-xs text-muted-foreground">
              {content.content_data_id}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
