import { useState, useRef, useMemo } from 'react'
import { Image, Upload, FileIcon, Folder, ChevronRight, X } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useMediaList, useMedia } from '@/queries/media'
import { useMediaUpload } from '@/hooks/use-media-upload'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { queryKeys } from '@/lib/query-keys'
import { nullStr, extractMediaPath, cn } from '@/lib/utils'
import type { FieldComponentProps } from '@/components/fields/field-renderer'

function isImageMimetype(mimetype: unknown): boolean {
  const mt = nullStr(mimetype)
  if (!mt) return false
  return mt.startsWith('image/')
}

function MediaPreview({ mediaId }: { mediaId: string }) {
  const { data: media } = useMedia(mediaId)
  if (!media) {
    return (
      <div className="flex items-center gap-2 rounded-md border px-3 py-2">
        <FileIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="truncate text-sm text-muted-foreground">{mediaId}</span>
      </div>
    )
  }
  return (
    <div className="flex items-center gap-3 rounded-md border px-3 py-2">
      {isImageMimetype(media.mimetype) ? (
        <img
          src={media.url}
          alt={nullStr(media.alt) || nullStr(media.display_name) || 'Media'}
          className="h-10 w-10 shrink-0 rounded object-cover"
        />
      ) : (
        <FileIcon className="h-10 w-10 shrink-0 text-muted-foreground/50" />
      )}
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">
          {nullStr(media.display_name) || nullStr(media.name) || 'Untitled'}
        </p>
        {nullStr(media.mimetype) && (
          <p className="text-xs text-muted-foreground">{nullStr(media.mimetype)}</p>
        )}
      </div>
    </div>
  )
}

function MediaPickerDialog({
  open,
  onOpenChange,
  onSelect,
  selectedId,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSelect: (mediaId: string) => void
  selectedId: string
}) {
  const { data: mediaList, isLoading } = useMediaList()
  const { upload, isUploading, progress, error: uploadError, reset } = useMediaUpload()
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [search, setSearch] = useState('')
  const [pathStack, setPathStack] = useState<string[]>([''])
  const [pendingId, setPendingId] = useState(selectedId)

  const currentPath = pathStack[pathStack.length - 1]

  function pushPath(path: string) {
    setPathStack((prev) => [...prev, path])
    setSearch('')
  }

  function popToIndex(stackIndex: number) {
    setPathStack((prev) => prev.slice(0, stackIndex + 1))
    setSearch('')
  }

  const { currentItems, subdirectories } = useMemo(() => {
    if (!mediaList) return { currentItems: [], subdirectories: [] }

    const dirSet = new Set<string>()
    const items: typeof mediaList = []

    for (const media of mediaList) {
      const { dir, filename } = extractMediaPath(media.url)
      const isDotfile = filename.startsWith('.')

      if (dir === currentPath) {
        if (!isDotfile) {
          items.push(media)
        }
      } else {
        const prefix = currentPath ? currentPath + '/' : ''
        if (dir.startsWith(prefix)) {
          const remaining = dir.slice(prefix.length)
          const nextSegment = remaining.split('/')[0]
          if (nextSegment) {
            dirSet.add(nextSegment)
          }
        }
      }
    }

    const sortedDirs = Array.from(dirSet).sort((a, b) =>
      a.localeCompare(b, undefined, { sensitivity: 'base' }),
    )

    return { currentItems: items, subdirectories: sortedDirs }
  }, [mediaList, currentPath])

  const filteredMedia = useMemo(() => {
    if (!search.trim()) return currentItems
    const term = search.toLowerCase()
    return currentItems.filter((m) => {
      const name = (nullStr(m.display_name) || nullStr(m.name) || '').toLowerCase()
      return name.includes(term)
    })
  }, [currentItems, search])

  const filteredDirs = useMemo(() => {
    if (!search.trim()) return subdirectories
    const term = search.toLowerCase()
    return subdirectories.filter((d) => d.toLowerCase().includes(term))
  }, [subdirectories, search])

  const pathSegments = currentPath ? currentPath.split('/') : []

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      const media = await upload(file, currentPath || undefined)
      queryClient.invalidateQueries({ queryKey: queryKeys.media.all })
      setPendingId(media.media_id)
    } catch {
      // error captured in hook state
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  function navigateToSegment(index: number) {
    const targetPath = index < 0 ? '' : pathSegments.slice(0, index + 1).join('/')
    const stackIdx = pathStack.lastIndexOf(targetPath)
    if (stackIdx !== -1) {
      popToIndex(stackIdx)
    } else {
      setPathStack(['', ...(targetPath ? [targetPath] : [])])
      setSearch('')
    }
  }

  function handleConfirm() {
    onSelect(pendingId)
    onOpenChange(false)
  }

  const hasContent = filteredDirs.length > 0 || filteredMedia.length > 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[80vh] flex-col sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>Select Media</DialogTitle>
        </DialogHeader>

        <div className="flex items-center gap-2">
          <Input
            placeholder="Search media..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1"
          />
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            onChange={handleFileChange}
          />
          <Button
            variant="outline"
            size="sm"
            onClick={() => fileInputRef.current?.click()}
            disabled={isUploading}
          >
            <Upload className="mr-1.5 h-3.5 w-3.5" />
            {isUploading ? `${progress}%` : 'Upload'}
          </Button>
        </div>

        {uploadError && (
          <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-sm text-destructive">
            Upload failed: {uploadError.message}
            <Button variant="ghost" size="sm" className="ml-2" onClick={reset}>
              Dismiss
            </Button>
          </div>
        )}

        {isUploading && (
          <div className="h-1.5 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
        )}

        {currentPath && (
          <div className="flex items-center gap-1 text-sm">
            <Button
              variant="ghost"
              size="sm"
              className="h-auto px-1 py-0.5"
              onClick={() => navigateToSegment(-1)}
            >
              Media
            </Button>
            {pathSegments.map((segment, i) => (
              <span key={i} className="flex items-center gap-1">
                <ChevronRight className="h-3 w-3 text-muted-foreground" />
                {i === pathSegments.length - 1 ? (
                  <span className="px-1 py-0.5 font-medium">{segment}</span>
                ) : (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-auto px-1 py-0.5"
                    onClick={() => navigateToSegment(i)}
                  >
                    {segment}
                  </Button>
                )}
              </span>
            ))}
          </div>
        )}

        <div className="flex-1 overflow-auto">
          {isLoading ? (
            <p className="py-8 text-center text-sm text-muted-foreground">Loading...</p>
          ) : hasContent ? (
            <div className="space-y-4">
              {filteredDirs.length > 0 && (
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                  {filteredDirs.map((dirName) => (
                    <button
                      key={'dir-' + dirName}
                      type="button"
                      className="flex items-center gap-3 rounded-md border border-border bg-card px-3 py-2.5 text-left transition-colors hover:bg-accent"
                      onClick={() =>
                        pushPath(currentPath ? currentPath + '/' + dirName : dirName)
                      }
                    >
                      <Folder className="h-5 w-5 shrink-0 text-muted-foreground" />
                      <span className="truncate text-sm font-medium">{dirName}</span>
                    </button>
                  ))}
                </div>
              )}

              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {filteredMedia.map((media) => {
                  const isSelected = media.media_id === pendingId
                  return (
                    <button
                      key={media.media_id}
                      type="button"
                      onClick={() => setPendingId(media.media_id)}
                      className="text-left"
                    >
                      <Card
                        className={cn(
                          'cursor-pointer overflow-hidden transition-colors',
                          isSelected
                            ? 'ring-2 ring-primary ring-offset-2 ring-offset-background'
                            : 'hover:border-primary/50',
                        )}
                      >
                        <div className="flex aspect-square items-center justify-center overflow-hidden bg-muted">
                          {isImageMimetype(media.mimetype) ? (
                            <img
                              src={media.url}
                              alt={
                                nullStr(media.alt) ||
                                nullStr(media.display_name) ||
                                nullStr(media.name) ||
                                'Media'
                              }
                              className="h-full w-full object-cover"
                            />
                          ) : (
                            <FileIcon className="h-10 w-10 text-muted-foreground/50" />
                          )}
                        </div>
                        <div className="space-y-1 p-2">
                          <p className="truncate text-xs font-medium">
                            {nullStr(media.display_name) ||
                              nullStr(media.name) ||
                              'Untitled'}
                          </p>
                          {nullStr(media.mimetype) && (
                            <Badge variant="secondary" className="text-[10px]">
                              {nullStr(media.mimetype)}
                            </Badge>
                          )}
                        </div>
                      </Card>
                    </button>
                  )
                })}
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12">
              <Image className="mb-3 h-10 w-10 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                {search.trim()
                  ? 'No media matches your search.'
                  : currentPath
                    ? 'This folder is empty.'
                    : 'No media yet. Upload a file to get started.'}
              </p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={!pendingId}>
            Select
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function MediaField({ field, value, onChange, error }: FieldComponentProps) {
  const [pickerOpen, setPickerOpen] = useState(false)

  function handleClear() {
    onChange('')
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={field.field_id}>{field.label}</Label>
      {value ? (
        <div className="space-y-2">
          <MediaPreview mediaId={value} />
          <div className="flex items-center gap-2">
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => setPickerOpen(true)}
            >
              Change
            </Button>
            <Button type="button" size="sm" variant="outline" onClick={handleClear}>
              <X className="mr-1 h-3 w-3" />
              Clear
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-2">
          <button
            type="button"
            onClick={() => setPickerOpen(true)}
            className="flex w-full items-center justify-center gap-2 rounded-md border border-dashed px-3 py-6 text-sm text-muted-foreground transition-colors hover:border-primary hover:text-foreground"
          >
            <Image className="h-4 w-4" />
            Select media
          </button>
        </div>
      )}
      {error && <p className="text-sm text-destructive">{error}</p>}
      {pickerOpen && (
        <MediaPickerDialog
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          onSelect={onChange}
          selectedId={value}
        />
      )}
    </div>
  )
}
