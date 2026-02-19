import { Plus } from 'lucide-react'
import type { Datatype } from '@modulacms/admin-sdk'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

type BlockInserterProps = {
  datatypes: Datatype[]
  onInsert: (datatypeId: string) => void
}

export function BlockInserter({ datatypes, onInsert }: BlockInserterProps) {
  const nonRoot = datatypes.filter((dt) => dt.type !== 'ROOT')

  if (nonRoot.length === 0) return null

  return (
    <div className="group relative flex items-center py-1">
      <div className="h-px flex-1 bg-transparent transition-colors group-hover:bg-border" />
      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            size="icon"
            className="h-6 w-6 rounded-full opacity-0 transition-opacity group-hover:opacity-100"
          >
            <Plus className="h-3 w-3" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-48 p-1" align="center">
          <div className="space-y-0.5">
            {nonRoot.map((dt) => (
              <button
                key={dt.datatype_id}
                className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                onClick={() => onInsert(dt.datatype_id)}
              >
                {dt.label}
              </button>
            ))}
          </div>
        </PopoverContent>
      </Popover>
      <div className="h-px flex-1 bg-transparent transition-colors group-hover:bg-border" />
    </div>
  )
}
