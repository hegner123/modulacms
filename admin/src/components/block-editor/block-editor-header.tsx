import { Link } from '@tanstack/react-router'
import type { ContentStatus } from '@modulacms/admin-sdk'
import { ArrowLeft, Settings, Save, Globe, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

type BlockEditorHeaderProps = {
  title: string
  slug: string
  rootStatus: ContentStatus
  dirtyCount: number
  saving: boolean
  publishing: boolean
  settingsOpen: boolean
  onToggleSettings: () => void
  onSaveAll: () => void
  onPublishAll: () => void
  onDraftAll: () => void
}

export function BlockEditorHeader({
  title,
  slug,
  rootStatus,
  dirtyCount,
  saving,
  publishing,
  settingsOpen,
  onToggleSettings,
  onSaveAll,
  onPublishAll,
  onDraftAll,
}: BlockEditorHeaderProps) {
  const isPublished = rootStatus === 'published'

  return (
    <div className="flex items-center gap-4 border-b px-6 py-3">
      <Button variant="ghost" size="sm" asChild>
        <Link to="/content">
          <ArrowLeft className="mr-1 h-4 w-4" />
          Back
        </Link>
      </Button>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <h1 className="text-lg font-semibold">{title}</h1>
          <Badge variant={isPublished ? 'default' : 'secondary'} className="text-xs">
            {rootStatus}
          </Badge>
          {dirtyCount > 0 && (
            <Badge variant="secondary" className="text-xs">
              {dirtyCount} unsaved
            </Badge>
          )}
        </div>
        <code className="text-xs text-muted-foreground">{slug}</code>
      </div>
      <div className="flex items-center gap-2">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              size="sm"
              variant={isPublished ? 'default' : 'outline'}
              disabled={publishing}
            >
              {isPublished ? (
                <Globe className="mr-1 h-4 w-4" />
              ) : (
                <FileText className="mr-1 h-4 w-4" />
              )}
              {publishing ? 'Updating...' : isPublished ? 'Published' : 'Draft'}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onPublishAll} disabled={publishing}>
              <Globe className="mr-2 h-4 w-4" />
              Publish All Blocks
            </DropdownMenuItem>
            <DropdownMenuItem onClick={onDraftAll} disabled={publishing}>
              <FileText className="mr-2 h-4 w-4" />
              Draft All Blocks
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          size="sm"
          onClick={onSaveAll}
          disabled={saving}
        >
          <Save className="mr-1 h-4 w-4" />
          {saving ? 'Saving...' : 'Save All'}
        </Button>
        <Button
          variant={settingsOpen ? 'secondary' : 'ghost'}
          size="icon"
          onClick={onToggleSettings}
        >
          <Settings className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
