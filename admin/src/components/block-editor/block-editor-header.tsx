import { Link } from '@tanstack/react-router'
import { ArrowLeft, Settings, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

type BlockEditorHeaderProps = {
  title: string
  slug: string
  dirtyCount: number
  saving: boolean
  settingsOpen: boolean
  onToggleSettings: () => void
  onSaveAll: () => void
}

export function BlockEditorHeader({
  title,
  slug,
  dirtyCount,
  saving,
  settingsOpen,
  onToggleSettings,
  onSaveAll,
}: BlockEditorHeaderProps) {
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
          {dirtyCount > 0 && (
            <Badge variant="secondary" className="text-xs">
              {dirtyCount} unsaved
            </Badge>
          )}
        </div>
        <code className="text-xs text-muted-foreground">{slug}</code>
      </div>
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          onClick={onSaveAll}
          disabled={dirtyCount === 0 || saving}
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
