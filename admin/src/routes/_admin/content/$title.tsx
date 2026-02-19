import { createFileRoute, Link } from '@tanstack/react-router'
import type { ContentTree } from '@modulacms/admin-sdk'
import { ArrowLeft, TreePine } from 'lucide-react'
import { useTree } from '@/queries/content'
import { useRoutes } from '@/queries/routes'
import { EmptyState } from '@/components/shared/empty-state'
import { BlockEditor } from '@/components/block-editor/block-editor'

export const Route = createFileRoute('/_admin/content/$title')({
  component: ContentEditorPage,
})

function ContentEditorPage() {
  const { title } = Route.useParams()
  const { data: routes, isLoading: routesLoading } = useRoutes()

  const decodedTitle = decodeURIComponent(title)
  const route = routes?.find((r) => r.title.toLowerCase() === decodedTitle)

  const { data: treeData, isLoading: treeLoading } = useTree(route?.slug ?? '')

  const isLoading = treeLoading || routesLoading

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (!route || !treeData) {
    return (
      <div className="space-y-4">
        <Link to="/content" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" />
          Back to Content
        </Link>
        <EmptyState
          icon={TreePine}
          title="Route not found"
          description="The content route could not be loaded."
        />
      </div>
    )
  }

  return <BlockEditor route={route} treeData={treeData as ContentTree} />
}
