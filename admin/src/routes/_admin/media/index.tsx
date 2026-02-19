import { useState, useRef, useMemo, useEffect } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import {
  Image,
  Upload,
  FileIcon,
  Folder,
  FolderPlus,
  ChevronRight,
} from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useMediaList } from '@/queries/media'
import { useMediaUpload } from '@/hooks/use-media-upload'
import { EmptyState } from '@/components/shared/empty-state'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import { nullStr, extractMediaPath } from '@/lib/utils'

export const Route = createFileRoute('/_admin/media/')({
  component: MediaLibraryPage,
})

// Persists across component mounts so path survives navigation to detail page and back
let persistedStack: string[] = ['']

function MediaLibraryPage() {
  const { data: mediaList, isLoading } = useMediaList()
  const {
    upload,
    isUploading,
    progress,
    error: uploadError,
    reset,
  } = useMediaUpload()
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [search, setSearch] = useState('')
  const [pathStack, setPathStack] = useState<string[]>(persistedStack)
  const [newFolderOpen, setNewFolderOpen] = useState(false)
  const [newFolderName, setNewFolderName] = useState('')
  const [creatingFolder, setCreatingFolder] = useState(false)

  const currentPath = pathStack[pathStack.length - 1]

  // Keep module-level stack in sync
  useEffect(() => {
    persistedStack = pathStack
  }, [pathStack])

  function pushPath(path: string) {
    setPathStack((prev) => [...prev, path])
    setSearch('')
  }

  function popToIndex(stackIndex: number) {
    setPathStack((prev) => prev.slice(0, stackIndex + 1))
    setSearch('')
  }

  // Derive items and subdirectories for the current path
  // Dotfiles (e.g. .keep) are used to persist empty directories but hidden from display
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
        // Check if this media is in a subdirectory of currentPath
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

  // Filter by search within current directory
  const filteredMedia = useMemo(() => {
    if (!search.trim()) return currentItems
    const term = search.toLowerCase()
    return currentItems.filter((m) => {
      const name = (
        nullStr(m.display_name) ||
        nullStr(m.name) ||
        ''
      ).toLowerCase()
      return name.includes(term)
    })
  }, [currentItems, search])

  // Filter subdirectories by search too
  const filteredDirs = useMemo(() => {
    if (!search.trim()) return subdirectories
    const term = search.toLowerCase()
    return subdirectories.filter((d) => d.toLowerCase().includes(term))
  }, [subdirectories, search])

  // Breadcrumb segments
  const pathSegments = currentPath ? currentPath.split('/') : []

  async function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      await upload(file, currentPath || undefined)
      queryClient.invalidateQueries({ queryKey: queryKeys.media.all })
    } catch {
      // error is captured in the hook state
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  async function handleCreateFolder() {
    const name = newFolderName.trim()
    if (!name) return
    const fullPath = currentPath ? currentPath + '/' + name : name
    setCreatingFolder(true)
    try {
      const dotfile = new File(['.'], '.keep', { type: 'application/octet-stream' })
      await upload(dotfile, fullPath)
      queryClient.invalidateQueries({ queryKey: queryKeys.media.all })
      pushPath(fullPath)
    } catch {
      // error is captured in the hook state
    } finally {
      setCreatingFolder(false)
    }
    setNewFolderName('')
    setNewFolderOpen(false)
  }

  function navigateToSegment(index: number) {
    // Find the stack entry that matches the target path
    const targetPath =
      index < 0 ? '' : pathSegments.slice(0, index + 1).join('/')
    const stackIdx = pathStack.lastIndexOf(targetPath)
    if (stackIdx !== -1) {
      popToIndex(stackIdx)
    } else {
      // Target wasn't in the stack history (e.g. deep-linked), reset to it
      setPathStack(['', ...(targetPath ? [targetPath] : [])])
      setSearch('')
    }
  }

  function isImageMimetype(mimetype: unknown): boolean {
    const mt = nullStr(mimetype)
    if (!mt) return false
    return mt.startsWith('image/')
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Media Library</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  const hasContent = filteredDirs.length > 0 || filteredMedia.length > 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Media Library</h1>
          <p className="text-muted-foreground">
            Upload, browse, and manage your media assets.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setNewFolderOpen(true)}
            title="New Folder"
          >
            <FolderPlus className="h-4 w-4" />
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            onChange={handleFileChange}
          />
          <Button
            onClick={() => fileInputRef.current?.click()}
            disabled={isUploading}
          >
            <Upload className="mr-2 h-4 w-4" />
            {isUploading ? `Uploading ${progress}%...` : 'Upload'}
          </Button>
        </div>
      </div>

      {uploadError && (
        <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
          Upload failed: {uploadError.message}
          <Button variant="ghost" size="sm" className="ml-2" onClick={reset}>
            Dismiss
          </Button>
        </div>
      )}

      {isUploading && (
        <div className="space-y-1">
          <div className="text-sm text-muted-foreground">Uploading...</div>
          <div className="h-2 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      <Input
        placeholder="Search media by name..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="max-w-sm"
      />

      {/* Breadcrumb navigation */}
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

      {/* Folder list */}
      {filteredDirs.length > 0 && (
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
          {filteredDirs.map((dirName) => (
            <button
              key={'dir-' + dirName}
              type="button"
              className="flex items-center gap-3 rounded-md border border-border bg-card px-3 py-2.5 text-left transition-colors hover:bg-accent"
              onClick={() =>
                pushPath(
                  currentPath ? currentPath + '/' + dirName : dirName,
                )
              }
            >
              <Folder className="h-5 w-5 shrink-0 text-muted-foreground" />
              <span className="truncate text-sm font-medium">{dirName}</span>
            </button>
          ))}
        </div>
      )}

      {hasContent ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {/* Media cards */}
          {filteredMedia.map((media) => (
            <Link
              key={media.media_id}
              to="/media/$id"
              params={{ id: media.media_id }}
            >
              <Card className="group cursor-pointer overflow-hidden transition-colors hover:border-primary/50">
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
                      className="h-full w-full object-cover transition-transform group-hover:scale-105"
                    />
                  ) : (
                    <FileIcon className="h-12 w-12 text-muted-foreground/50" />
                  )}
                </div>
                <div className="space-y-1 p-3">
                  <p className="truncate text-sm font-medium">
                    {nullStr(media.display_name) ||
                      nullStr(media.name) ||
                      'Untitled'}
                  </p>
                  {nullStr(media.mimetype) && (
                    <Badge variant="secondary" className="text-xs">
                      {nullStr(media.mimetype)}
                    </Badge>
                  )}
                </div>
              </Card>
            </Link>
          ))}
        </div>
      ) : (
        <EmptyState
          icon={Image}
          title="No media found"
          description={
            search.trim()
              ? 'No media assets match your search.'
              : currentPath
                ? 'This folder is empty. Upload a file or create a subfolder.'
                : 'Upload your first media asset to get started.'
          }
          action={
            !search.trim() ? (
              <Button
                onClick={() => fileInputRef.current?.click()}
                disabled={isUploading}
              >
                <Upload className="mr-2 h-4 w-4" />
                Upload
              </Button>
            ) : undefined
          }
        />
      )}

      {/* New Folder Dialog */}
      <Dialog open={newFolderOpen} onOpenChange={setNewFolderOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>New Folder</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            {currentPath && (
              <p className="text-sm text-muted-foreground">
                Creating in: {currentPath}/
              </p>
            )}
            <Input
              placeholder="Folder name"
              value={newFolderName}
              onChange={(e) => setNewFolderName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreateFolder()
              }}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setNewFolderName('')
                setNewFolderOpen(false)
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateFolder}
              disabled={!newFolderName.trim() || creatingFolder}
            >
              {creatingFolder ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
