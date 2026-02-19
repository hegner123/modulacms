import { useState, useMemo } from 'react'
import type { ContentTree, ContentNode, ContentID, DatatypeID, UserID, ContentStatus, Route } from '@modulacms/admin-sdk'
import { TreePine } from 'lucide-react'
import { useCreateContentData } from '@/queries/content'
import { useDatatypes, useDatatypeFieldsByDatatype } from '@/queries/datatypes'
import { useFields } from '@/queries/fields'
import { useAuthContext } from '@/lib/auth'
import { EmptyState } from '@/components/shared/empty-state'
import { BlockEditorHeader } from './block-editor-header'
import { PageFields } from './page-fields'
import { BlockList } from './block-list'
import { BlockInserter } from './block-inserter'
import { DocumentSettings } from './document-settings'
import { useBlockEditorState } from './use-block-editor-state'
import { buildMergedFields } from './build-merged-fields'

type BlockEditorProps = {
  route: Route
  treeData: ContentTree
}

function RootFieldsContainer({
  rootNode,
  state,
}: {
  rootNode: ContentNode
  state: ReturnType<typeof useBlockEditorState>
}) {
  const datatypeId = rootNode.datatype.info.datatype_id
  const { data: datatypeFields } = useDatatypeFieldsByDatatype(datatypeId)
  const { data: allFields } = useFields()

  const mergedFields = useMemo(
    () => buildMergedFields(rootNode, datatypeFields, allFields),
    [rootNode, datatypeFields, allFields],
  )

  return (
    <PageFields
      node={rootNode}
      mergedFields={mergedFields}
      getFieldValue={state.getFieldValue}
      setFieldValue={state.setFieldValue}
    />
  )
}

function countAllBlocks(node: ContentNode): number {
  const children = node.nodes ?? []
  let count = children.length
  for (const child of children) {
    count += countAllBlocks(child)
  }
  return count
}

export function BlockEditor({ route, treeData }: BlockEditorProps) {
  const { user } = useAuthContext()
  const { data: datatypes } = useDatatypes()
  const createContent = useCreateContentData()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const state = useBlockEditorState()

  const tree = treeData as ContentTree
  const rootNode = tree.root
  const childNodes = rootNode?.nodes ?? []

  if (!rootNode) {
    return (
      <EmptyState
        icon={TreePine}
        title="No content"
        description="This route has no content data yet."
      />
    )
  }

  function handleInsert(parentId: string, datatypeId: string) {
    const now = new Date().toISOString()
    createContent.mutate({
      route_id: route.route_id,
      parent_id: parentId as unknown as ContentID,
      first_child_id: null,
      next_sibling_id: null,
      prev_sibling_id: null,
      datatype_id: datatypeId as DatatypeID,
      author_id: (user?.user_id ?? null) as UserID | null,
      status: 'draft' as ContentStatus,
      date_created: now,
      date_modified: now,
    })
  }

  const rootContentDataId = rootNode.datatype.content.content_data_id

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col">
      <BlockEditorHeader
        title={route.title}
        slug={route.slug}
        dirtyCount={0}
        saving={state.saving}
        settingsOpen={settingsOpen}
        onToggleSettings={() => setSettingsOpen(!settingsOpen)}
        onSaveAll={() => {}}
      />

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-auto">
          <div className="mx-auto max-w-3xl space-y-6 px-6 py-8">
            <RootFieldsContainer rootNode={rootNode} state={state} />

            {childNodes.length > 0 ? (
              <BlockList
                childNodes={childNodes}
                datatypes={datatypes ?? []}
                state={state}
                onInsert={handleInsert}
                parentId={rootContentDataId}
              />
            ) : (
              <div className="py-8 text-center">
                <p className="mb-4 text-sm text-muted-foreground">
                  No content blocks yet. Use the + button to add one.
                </p>
                <BlockInserter
                  datatypes={datatypes ?? []}
                  onInsert={(dtId) => handleInsert(rootContentDataId, dtId)}
                />
              </div>
            )}
          </div>
        </div>

        {settingsOpen && (
          <DocumentSettings
            rootNode={rootNode}
            blockCount={countAllBlocks(rootNode)}
          />
        )}
      </div>
    </div>
  )
}
