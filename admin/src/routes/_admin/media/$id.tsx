import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { ArrowLeft, FileIcon, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { useMedia, useUpdateMedia, useDeleteMedia } from '@/queries/media'
import { ConfirmDialog } from '@/components/shared/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Separator } from '@/components/ui/separator'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { formatDateTime, nullStr } from '@/lib/utils'

export const Route = createFileRoute('/_admin/media/$id')({
  component: MediaDetailPage,
})

type EditFormValues = {
  display_name: string
  alt: string
  caption: string
  description: string
}

function MediaDetailPage() {
  const { id } = Route.useParams()
  const navigate = useNavigate()
  const { data: media, isLoading } = useMedia(id)
  const updateMedia = useUpdateMedia()
  const deleteMedia = useDeleteMedia()

  const [deleteOpen, setDeleteOpen] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { isDirty },
    reset,
  } = useForm<EditFormValues>({
    values: media
      ? {
          display_name: nullStr(media.display_name) || '',
          alt: nullStr(media.alt) || '',
          caption: nullStr(media.caption) || '',
          description: nullStr(media.description) || '',
        }
      : {
          display_name: '',
          alt: '',
          caption: '',
          description: '',
        },
  })

  function onSave(values: EditFormValues) {
    if (!media) return
    updateMedia.mutate(
      {
        media_id: media.media_id,
        name: media.name,
        display_name: values.display_name || null,
        alt: values.alt || null,
        caption: values.caption || null,
        description: values.description || null,
        class: media.class,
        mimetype: media.mimetype,
        dimensions: media.dimensions,
        url: media.url,
        srcset: media.srcset,
        author_id: media.author_id,
        date_created: media.date_created,
        date_modified: new Date().toISOString(),
      },
      {
        onSuccess: (updated) => {
          reset({
            display_name: nullStr(updated.display_name) || '',
            alt: nullStr(updated.alt) || '',
            caption: nullStr(updated.caption) || '',
            description: nullStr(updated.description) || '',
          })
        },
      },
    )
  }

  function handleDelete() {
    if (!media) return
    deleteMedia.mutate(media.media_id, {
      onSuccess: () => {
        navigate({ to: '/media' })
      },
    })
  }

  function isImageMimetype(mimetype: unknown): boolean {
    const mt = nullStr(mimetype)
    if (!mt) return false
    return mt.startsWith('image/')
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Media Detail</h1>
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (!media) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Media Not Found</h1>
        <p className="text-muted-foreground">
          The requested media asset could not be found.
        </p>
        <Button variant="outline" asChild>
          <Link to="/media">Back to Media</Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link to="/media">
            <ArrowLeft className="h-4 w-4" />
          </Link>
        </Button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">
            {nullStr(media.display_name) || nullStr(media.name) || 'Untitled Media'}
          </h1>
          <p className="text-sm text-muted-foreground">
            Edit media details and metadata.
          </p>
        </div>
        <Button
          variant="destructive"
          size="sm"
          onClick={() => setDeleteOpen(true)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Preview</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center overflow-hidden rounded-lg bg-muted">
              {isImageMimetype(media.mimetype) ? (
                <img
                  src={media.url}
                  alt={nullStr(media.alt) || nullStr(media.display_name) || nullStr(media.name) || 'Media'}
                  className="max-h-96 w-full object-contain"
                />
              ) : (
                <div className="flex flex-col items-center gap-2 py-16">
                  <FileIcon className="h-16 w-16 text-muted-foreground/50" />
                  <p className="text-sm text-muted-foreground">
                    {nullStr(media.mimetype) || 'Unknown type'}
                  </p>
                </div>
              )}
            </div>

            <Separator className="my-4" />

            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Mimetype</span>
                <span>{nullStr(media.mimetype) || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">URL</span>
                <a
                  href={media.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="max-w-[200px] truncate text-primary hover:underline"
                >
                  {media.url}
                </a>
              </div>
              {nullStr(media.srcset) && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Srcset</span>
                  <span className="max-w-[200px] truncate">{nullStr(media.srcset)}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span>{formatDateTime(media.date_created)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Modified</span>
                <span>{formatDateTime(media.date_modified)}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Details</CardTitle>
            <CardDescription>
              Update the display name, alt text, caption, and description.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSave)} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-display-name">Display Name</Label>
                <Input
                  id="edit-display-name"
                  placeholder="Display name"
                  {...register('display_name')}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-alt">Alt Text</Label>
                <Input
                  id="edit-alt"
                  placeholder="Descriptive alt text"
                  {...register('alt')}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-caption">Caption</Label>
                <Input
                  id="edit-caption"
                  placeholder="Caption"
                  {...register('caption')}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="edit-description">Description</Label>
                <Textarea
                  id="edit-description"
                  placeholder="Detailed description"
                  rows={4}
                  {...register('description')}
                />
              </div>

              <Button
                type="submit"
                disabled={!isDirty || updateMedia.isPending}
              >
                {updateMedia.isPending ? 'Saving...' : 'Save Changes'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Delete Media"
        description={`Are you sure you want to delete "${nullStr(media.display_name) || nullStr(media.name) || 'this media'}"? This action cannot be undone.`}
        onConfirm={handleDelete}
        loading={deleteMedia.isPending}
        variant="destructive"
      />
    </div>
  )
}
